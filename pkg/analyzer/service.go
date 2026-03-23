package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/atomic"

	conf "smart-slowquery/conf/analyzer"
	mysql2 "smart-slowquery/internal/model/mysql"
	timeUtil "smart-slowquery/internal/util/time"
	"smart-slowquery/pkg/cache"
	sqlFinger "smart-slowquery/pkg/finger/mysql"
	sysMetrics "smart-slowquery/pkg/metrics/analyzer"
	"smart-slowquery/pkg/service/dbms"
	storeRequest "smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"

	"smart-slowquery/internal/model/filebeat"
	"smart-slowquery/internal/util/hint"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/parser/mysql"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/clickhouse"
	myysqlCli "smart-slowquery/pkg/store/mysql"

	"github.com/Shopify/sarama"
	"github.com/mitchellh/mapstructure"
)

const (
	historyFlag = 0
	newFlag     = 1

	noHitIndex  = 0
	hasHitIndex = 1

	closeSwitch = 0
	openSwitch  = 1
)

var filterUser = map[string]struct{}{
	"statssvr":      {},
	"dba_meta":      {},
	"shopeeroot":    {},
	"rds_monitor":   {},
	"shopee":        {},
	"db_archive_rw": {},
}
var filterDB = map[string]struct{}{
	"information_schema": {},
	"mysql":              {},
	"sys":                {},
	"test":               {},
}

type Service struct {
	finger      *sqlFinger.FingerPrint
	parser      *mysql.SlowLogParser
	querys      []*filebeat.SlowQuery
	writer      store.CKWriter
	reader      *clickhouse.Client
	monitorDB   myysqlCli.DB
	monitorData atomic.Value
	adminUser   string
	adminPass   string
	dbmsSrv     *dbms.DBMetaService
	cfg         *conf.Config
}

func NewService(cfg *conf.Config, keepHint bool, ckCli *clickhouse.Client, monitorDB myysqlCli.DB, dbmsSrv *dbms.DBMetaService) (srv *Service, err error) {
	if cfg.Analyzer.Pattern == nil {
		return nil, fmt.Errorf("pattern is empty")
	}
	var parser *mysql.SlowLogParser
	if parser, err = mysql.NewSlowLogParser(cfg.Analyzer.Pattern); err != nil {
		return nil, err
	}
	service := &Service{
		finger: sqlFinger.NewFingerPrint(keepHint),
		writer: clickhouse.NewFlusher(ckCli, cfg.CKFlush),
		reader: ckCli,
		parser: parser,
		// fingerFilter:  fingerFilter,
		monitorDB:   monitorDB,
		monitorData: atomic.Value{},
		adminUser:   cfg.MysqlAccessConfig.User,
		adminPass:   cfg.MysqlAccessConfig.Key,
		dbmsSrv:     dbmsSrv,
		cfg:         cfg,
	}
	log.Info("according db name load dbms cmdb info start ...")
	service.getDBMSInfoToCache()
	log.Info("according db name load dbms cmdb info over ...")

	data, err := service.fetch()
	if err != nil {
		return nil, err
	}
	service.monitorData.Store(data)
	go service.alwaysFetch()
	return service, nil
}

func (srv *Service) getDBMSInfoToCache() {
	dbList, err := srv.reader.GetDBNameList()
	if err != nil {
		log.Errorf("getDBMSInfoToCache error:%s", err.Error())
	}
	for _, dbName := range dbList {
		l1l2, team, roleMap, err := srv.dbmsSrv.GetL1L2AndTeamAndRoleInfo(dbName, srv.cfg.Basic.ENV)
		if err != nil {
			log.Errorf("GetL1L2AndTeamAndRoleInfo error:%s", err.Error())
			continue
		}
		log.Infof("GetL1L2AndTeamAndRoleInfo dbname:%s l1l2:%s,team:%s", dbName, l1l2, team)
		cache.StoreData(dbName, cache.DBMSInfo{DBName: dbName, L1L2: l1l2, Team: team, RoleMap: roleMap})
	}
}

func (srv *Service) alwaysFetch() {
	ticket := time.NewTicker(time.Second * 60)
	defer ticket.Stop()
	for range ticket.C {
		data, err := srv.fetch()
		if err != nil {
			log.Errorf("fetch error:%s", err.Error())
			continue
		}
		log.Infof("alwaysFetch load %d data\n", len(data))
		srv.monitorData.Store(data)
	}
}

func (srv *Service) getL1L2ByDBName(databaseName, env string) (string, string, func(string) string) {
	result, exist := cache.FeetchData(databaseName)
	if !exist {
		// fetch from dbms && set key/value
		l1l2, team, roleMap, err := srv.dbmsSrv.GetL1L2AndTeamAndRoleInfo(databaseName, env)
		if err != nil {
			log.Errorf("fetch dbms l1l2 detail error:%s", err.Error())
			return "", "", nil
		}
		cache.StoreData(databaseName, cache.DBMSInfo{DBName: databaseName, L1L2: l1l2, Team: team, RoleMap: roleMap})
		return l1l2, team, func(s string) string {
			if _, ok := roleMap[s]; ok {
				return roleMap[s]
			}
			return "Unknown"
		}
	}
	return result.L1L2, result.DBName, func(s string) string {
		if _, ok := result.RoleMap[s]; ok {
			return result.RoleMap[s]
		}
		return "Unknown"
	}
}

func (srv *Service) matchNameAndTime(hostIP string, logTime time.Time) (matched bool, err error) {
	logTime = timeUtil.ConvertToBaseDate(logTime)
	m := srv.monitorData.Load().(map[string]*mysql2.DBFreeSizeTab)
	data, ok := m[hostIP]
	if !ok {
		return false, nil
	}
	start, err := timeUtil.HourMinuteSecond(data.BottomStart)
	if err != nil {
		return false, err
	}
	end, err := timeUtil.HourMinuteSecond(data.BottomEnd)
	if err != nil {
		return false, err
	}
	if logTime.After(start) && logTime.Before(end) {
		log.Infof("%s ip matched,current time is %s, db time is :%s-%s", hostIP, logTime.String(), start.String(), end.String())
		return true, nil
	}
	return false, nil
}

func (srv *Service) isHitIndex(host, dbName, sql string, port int32) (bool, error) {
	explainResult, err := srv.execExplain(host, dbName, sql, port)
	if err != nil {
		return false, err
	}
	if err = explainResult.GetExplainError(); err != nil {
		return false, err
	}
	for _, re := range explainResult.GetExplainInfos() {
		// key字段是空
		if re.Key.String == "NULL" || re.Key.String == "" {
			return false, nil
		}
	}
	// key字段不是空就是命中索引
	return true, nil
}

func (srv *Service) execExplain(host, dbName, sql string, port int32) (*response.ExplainResult, error) {
	var (
		err     error
		s       *myysqlCli.Session
		results []*response.ExplainInfo
	)

	if s, err = myysqlCli.NewSession(srv.adminUser, srv.adminPass, dbName, host, int32(port)); err != nil {
		log.Errorf("fetchExplain NewSession error:%s", err.Error())
		_ = s.Close()
		return nil, err
	}
	// 判断五秒超时
	if results, err = s.ExplainWithTimeout(sql, 5*time.Second); err != nil {
		log.Errorf("fetchExplain ExplainWithTimeout statement:%s ,session:%s,error:%s", sql, s.String(), err.Error())
	}
	_ = s.Close()
	if len(results) == 0 {
		return nil, errors.New("no explain result")
	}
	return response.NewExplainResult(results, err), err
}

func (srv *Service) fetch() (dataMap map[string]*mysql2.DBFreeSizeTab, err error) {
	data := make([]*mysql2.DBFreeSizeTab, 0)
	dataMap = make(map[string]*mysql2.DBFreeSizeTab)
	if err = srv.monitorDB.Query(&data, "ts = ? and bottom_start != '' and bottom_end != ''", timeUtil.GetToday20AMTimestamp()); err != nil {
		return nil, err
	}
	log.Infof("first load %d data\n", len(data))
	for _, v := range data {
		dataMap[v.IP] = v
	}
	return dataMap, nil
}

func (srv *Service) Processor(msg *sarama.ConsumerMessage) (err error, filter, flushed bool) {
	var (
		fb    *filebeat.SlowQuery
		start = time.Now()
	)
	//最后运行
	defer func() {
		sysMetrics.CollectServiceMetrics("Processor", sysMetrics.GetStatus(err), time.Since(start))
	}()
	//buildFileBeat会返回一个fb 包含了慢查询的信息：集群，DB，SQL句子...
	if fb, err = srv.buildFileBeat(string(msg.Value)); err != nil || !fb.Valid() {
		sysMetrics.CollectMessageFilteredCounter(fmt.Sprintf("analyzer_%s_filtered", msg.Topic), sysMetrics.GetStatus(err), 1)
		log.Debugf("Processor topic:%s,partition:%d,offset:%d,message:%s,logTime:%d, error:%v", msg.Topic, msg.Partition, msg.Offset, string(msg.Value), msg.Timestamp.Unix(), err)
		return nil, true, false
	}
	// 现在得到的fb 里面包含了慢sql的cluster dbname sql信息,过滤掉不需要进行存储分析的SQL
	// 过滤insert into的慢日志
	if strings.Contains(fb.SlowLog.Query, "insert into") {
		return nil, true, false
	}
	// 过滤系统用户的慢日志
	if _, ok := filterUser[fb.SlowLog.User]; ok {
		return nil, true, false
	}
	// 过滤系统db
	if _, ok := filterDB[fb.SlowLog.CurrentDB]; ok {
		return nil, true, false
	}

	// query is_hit_index
	var matched bool // 看是否在低峰时间段 是的话才进行EXPLAIN连数据库做分析否则只是采集慢日志
	matched, err = srv.matchNameAndTime(fb.Fields.InstanceHost, time.Unix(fb.SlowLog.TimeStamp, 0))
	if err != nil {
		log.Errorf("Processor message topic:%s,partition:%d,offset:%d,logTime:%d matchNameAndTime error:%v", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp.Unix(), err)
	}
	log.Infof("%s ip matched, db:%s", fb.Fields.InstanceHost, fb.SlowLog.CurrentDB)

	// 查询执行计划，确定是否命中索引
	fb.HasOpenIndexSwitch = closeSwitch // 0 = 未开启EXPLAIN检查
	fb.IsHitIndex = noHitIndex          // 0 = 未命中索引
	if matched {                        // 在低峰阶段
		fb.HasOpenIndexSwitch = openSwitch // 开启EXPLAIN检查
		if has, err := srv.isHitIndex(fb.Fields.InstanceHost, fb.SlowLog.CurrentDB, fb.SlowLog.Query, fb.Fields.InstancePort); err != nil {
			log.Errorf("Processor message topic:%s,partition:%d,offset:%d,logTime:%d isHitIndex error:%v", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp.Unix(), err)
		} else {
			if has {
				fb.IsHitIndex = hasHitIndex
			}
		} // if - else 的逻辑是if err ！= nil（执行出错） else 执行没问题 if has has的意思是命中索引，执行没问题就把命中索引给他加上去
	}
	// 获取团队信息 非核心
	var roleFunc func(string) string
	fb.L1L2, fb.Team, roleFunc = srv.getL1L2ByDBName(fb.SlowLog.CurrentDB, fb.Fields.Env)

	if roleFunc != nil {
		fb.Role = roleFunc(fb.Fields.InstanceHost)
	}
	// 批量写入 如果不够缓存区没有1000条不会写入到clickhouse
	if err, flushed = srv.writer.Append(storeRequest.BuildSlowQueryLog(fb)); err != nil {
		log.Errorf("Processor message topic:%s,partition:%d,offset:%d,logTime:%d writer Append, error:%v", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp.Unix(), err)
		return err, false, flushed
	}
	// filter 被过滤是因为insert,dba团队人员的SQL
	log.Infof("Processor message topic:%s,partition:%d,offset:%d finish", msg.Topic, msg.Partition, msg.Offset)
	return nil, false, flushed
}

// Go里面返回的fb和err 已经在方法名定义好了 内部可以直接使用
func (srv *Service) buildFileBeat(message string) (fb *filebeat.SlowQuery, err error) {
	if len(message) == 0 {
		return nil, fmt.Errorf("message is empty")
	}

	var (
		infos   map[string]interface{} // map[string]interface{}中的interface{}不是传统意义上用于定义行为规范的接口，而是Go语言提供的一种灵活存储任意类型数据的机制
		slowLog filebeat.MysqlSlowLog
	)
	// 反序列化从kafka拿到的慢日志 失败了返回空的fb和err
	// ？这里传递了&fb fb是一个指针 指针的指针作用为：函数意识到这是个空指针就会建立一个空结构体实例填充字段修改fb的指向
	if err = json.Unmarshal([]byte(message), &fb); err != nil {
		return nil, err
	}
	// 把从kafka拿到的message按照字段填充到infos-这个infos的类型是一个map interface可以适配各种类型的value
	if infos, err = srv.parser.ParserSlowLog(fb.Message); err != nil {
		return nil, err
	}
	// 把填充好的infos 填充到下面流程可以使用的结构体之中 mapstructure是导入的包
	if err = mapstructure.Decode(infos, &slowLog); err != nil {
		return nil, err
	}
	//TODO：为什么走了这么多步-> （答案写在飞书云文档）为什么中间弄出来一个infos和slowlog 开始反序列化的时候就给fb做填充了 然后把fb的Message（string类型）拿出来 填充到infos 然后把infos填充到slowlog 最后把slowlog填充到fb 为啥这么复杂？每个阶段的执行结果是怎么样的
	fb.SlowLog = &slowLog
	fb.SlowLog.Hint = hint.GetSqlTraceHint(fb.SlowLog.Query) // Hint的意思是暗示，这里获取trace id（可以定位业务请求，这id不是本项目生成的）

	// 解决分库场景下，查询语句中缺少db名称导致fingerID重复的问题
	_, digest := srv.finger.NormalizeDigest(fmt.Sprintf("%s-%s-%s", fb.SlowLog.Query, fb.Fields.ClusterUUID, fb.SlowLog.CurrentDB))
	fb.SlowLog.FingerID = digest.String()
	fb.SlowLog.FingerSql, _ = srv.finger.NormalizeDigest(fb.SlowLog.Query)
	// 同时保留了FingerSql 分类依据为sql 不同的数据库相同sql会聚合在一起

	return fb, err
}

func (srv *Service) Close() {
	if err := srv.writer.FlushAll(); err != nil {
		fmt.Printf("close flush all error:%s \n", err.Error())
		return
	}
}

// GetWriter 获取 ClickHouse 写入器，用于批量处理器
func (srv *Service) GetWriter() store.CKWriter {
	return srv.writer
}

// ProcessSingleMessage 处理单条消息，返回处理后的数据和过滤状态
// 这个方法与 Processor 类似，但不进行写入操作，只返回处理后的数据
func (srv *Service) ProcessSingleMessage(msg *sarama.ConsumerMessage) (*filebeat.SlowQuery, error, bool) {
	var (
		fb    *filebeat.SlowQuery
		start = time.Now()
		err   error
	)
	
	// 最后运行
	defer func() {
		sysMetrics.CollectServiceMetrics("ProcessSingleMessage", sysMetrics.GetStatus(err), time.Since(start))
	}()
	
	// buildFileBeat会返回一个fb 包含了慢查询的信息：集群，DB，SQL句子...
	if fb, err = srv.buildFileBeat(string(msg.Value)); err != nil || !fb.Valid() {
		sysMetrics.CollectMessageFilteredCounter(fmt.Sprintf("analyzer_%s_filtered", msg.Topic), sysMetrics.GetStatus(err), 1)
		log.Debugf("ProcessSingleMessage topic:%s,partition:%d,offset:%d,message:%s,logTime:%d, error:%v", msg.Topic, msg.Partition, msg.Offset, string(msg.Value), msg.Timestamp.Unix(), err)
		return nil, nil, true
	}
	
	// 过滤insert into的慢日志
	if strings.Contains(fb.SlowLog.Query, "insert into") {
		return nil, nil, true
	}
	
	// 过滤系统用户的慢日志
	if _, ok := filterUser[fb.SlowLog.User]; ok {
		return nil, nil, true
	}
	
	// 过滤系统db
	if _, ok := filterDB[fb.SlowLog.CurrentDB]; ok {
		return nil, nil, true
	}

	// query is_hit_index
	var matched bool // 看是否在低峰时间段 是的话才进行EXPLAIN连数据库做分析否则只是采集慢日志
	matched, err = srv.matchNameAndTime(fb.Fields.InstanceHost, time.Unix(fb.SlowLog.TimeStamp, 0))
	if err != nil {
		log.Errorf("ProcessSingleMessage message topic:%s,partition:%d,offset:%d,logTime:%d matchNameAndTime error:%v", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp.Unix(), err)
	}
	log.Infof("%s ip matched, db:%s", fb.Fields.InstanceHost, fb.SlowLog.CurrentDB)

	// 查询执行计划，确定是否命中索引
	fb.HasOpenIndexSwitch = closeSwitch // 0 = 未开启EXPLAIN检查
	fb.IsHitIndex = noHitIndex          // 0 = 未命中索引
	if matched {                        // 在低峰阶段
		fb.HasOpenIndexSwitch = openSwitch // 开启EXPLAIN检查
		if has, err := srv.isHitIndex(fb.Fields.InstanceHost, fb.SlowLog.CurrentDB, fb.SlowLog.Query, fb.Fields.InstancePort); err != nil {
			log.Errorf("ProcessSingleMessage message topic:%s,partition:%d,offset:%d,logTime:%d isHitIndex error:%v", msg.Topic, msg.Partition, msg.Offset, msg.Timestamp.Unix(), err)
		} else {
			if has {
				fb.IsHitIndex = hasHitIndex
			}
		} // if - else 的逻辑是if err ！= nil（执行出错） else 执行没问题 if has has的意思是命中索引，执行没问题就把命中索引给他加上去
	}
	
	// 获取团队信息 非核心
	var roleFunc func(string) string
	fb.L1L2, fb.Team, roleFunc = srv.getL1L2ByDBName(fb.SlowLog.CurrentDB, fb.Fields.Env)

	if roleFunc != nil {
		fb.Role = roleFunc(fb.Fields.InstanceHost)
	}
	
	log.Infof("ProcessSingleMessage message topic:%s,partition:%d,offset:%d finish", msg.Topic, msg.Partition, msg.Offset)
	return fb, nil, false
}
