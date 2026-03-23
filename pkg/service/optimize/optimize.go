package optimize

import (
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	mysqlParser "smart-slowquery/pkg/parser/mysql"
	cmdbService "smart-slowquery/pkg/service/cmdb"
	storeMysql "smart-slowquery/pkg/store/mysql"

	"smart-slowquery/internal/util/hint"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/log"
	"smart-slowquery/thrid-party/soar-dev/advisor"
	"smart-slowquery/thrid-party/soar-dev/ast"
	"smart-slowquery/thrid-party/soar-dev/common"
	"smart-slowquery/thrid-party/soar-dev/database"

	"fmt"
	"time"

	"vitess.io/vitess/go/vt/sqlparser"
)

type Service struct{}

// 这些都是指针接收器 不带*是值接收器
func NewService() (*Service, error) {
	return &Service{}, nil
}

// GetSuggests
// mysqlSuggest 语法检查建议 优先级最高 出错直接不建议了
// heuristicSuggest写法检查 性能问题（全表扫描） 安全问题（delete没有where）索引失效 优先级次之 还是在改SQL TODO
// indexSuggest索引优化，依据表结构给出索引建议 具体的CREATE INDEX 这是在加索引 优先级最低 因为需要占用空间 需要DBA审批
// 优化破sql比盲目加索引重要
// /*
func (s *Service) GetSuggests(param *request.OptimizeRequest, traceID string) (mysqlSuggest, indexSuggest, heuristicSuggest map[string]advisor.Rule, err error) {
	var (
		tableInfo  map[string]*database.TableInfo
		idxAdvisor *advisor.IndexAdvisor
	)

	start := time.Now()
	defer func() {
		// 当启发式建议存在，索引优化不存在的时候，提示去进行启发式建议优化
		//Q1:返回的三个建议现在还没有生产 heuristicSuggest不是肯定是0？：No 遇到defer 先注册这个函数但不执行 在return之前执行这个函数内部内容
		if len(heuristicSuggest) != 0 && len(indexSuggest) == 0 {
			//填充...
			indexSuggest = map[string]advisor.Rule{"CHECK.006": {
				Item:    "CHECK.006",
				Summary: "No suggestions，Please view the「Sql Audit」 tag on the left",
				Content: "There are no index optimization suggestions. You can check 「Sql Audit」 to optimize your statements.",
			}}
		}
		// 合并冲突规则 合并重复的建议
		s.formatSuggest(param.SQL, param.CurrentDB, mysqlSuggest, indexSuggest, heuristicSuggest)
		sysMetrics.CollectServiceMetrics("Service.GetSuggests", sysMetrics.GetStatus(err), time.Since(start))
	}()
	//最先执行的是这里
	sqlType, parseErr := mysqlParser.ParseSqlStatement(hint.RemoveHint(param.SQL))
	if parseErr != nil { //解析失败了 先看下整体sql是什么类型的 这里使用了tidb parser解析
		log.Errorf("get statement type err, sql :%s,error:%s", param.SQL, parseErr.Error())
		errResult := advisor.RuleMySQLError("ERR.000", parseErr)
		errResult.Case = param.SQL
		mysqlSuggest = map[string]advisor.Rule{"ERR.000": errResult}
		return
	}

	if sqlType == mysqlParser.InsertStatement { //Insert语句慢可能是磁盘io/锁竞争，非索引方向
		log.Errorf("not support insert sql to exec explain, sql:%s,", param.SQL)
		heuristicSuggest = map[string]advisor.Rule{"CHECK.000": {
			Item:    "CHECK.000",
			Summary: "The insert statement does not require index optimization",
			Content: "The insert statement does not require index optimization",
		}}
		return
	}

	// 语法树解析，如果语法检查出错则不需要给优化建议 里面有两个语法树结果 这里调用了soar
	// 分两层解析是为了不合格的直接去掉 第二次解析更复杂消耗cpu
	queryAudit, syntaxErr := advisor.NewQuery4Audit(param.SQL, traceID)

	if syntaxErr != nil { //解析出错
		// tidb parser 语法检查给出的建议 ERR.000
		errResult := advisor.RuleMySQLError("ERR.000", syntaxErr)
		errResult.Case = param.SQL
		mysqlSuggest = map[string]advisor.Rule{"ERR.000": errResult}
		return
	}
	//返回true 检验是不是非DML和DQL（select）如果是 返回ture 是true 进入if内容 如果是dql和dml不会进入到if 直接到dbmete那了
	if sure, description := isNotDMLORDQL(queryAudit); sure {
		heuristicSuggest = map[string]advisor.Rule{"CHECK.002": {
			Item:    "CHECK.002",
			Summary: description,
			Content: description,
		}}
		return
	}
	//从SQL语法树中提取数据库和表的元信息 比如数据库的名字 表的字段这些不是实际的信息内容 底层逻辑先不看
	dbMeta := ast.GetMeta(queryAudit.Stmt, nil).SetDefault(param.CurrentDB)

	// 不支持垮库(sql语句)查询
	if isCrossDatabase(dbMeta, param.CurrentDB) {
		heuristicSuggest = map[string]advisor.Rule{"CHECK.004": advisor.Rule{
			Item:    "CHECK.004",
			Summary: "cross-database query is not supported",
			Content: "cross-database query is not supported",
		}}
		return
	}

	// 查询sql语句所属线上表的相关信息 tableInfo表信息，后续sql优化需要这个
	tableInfo, err = fetchTableInfo(param.User, param.Pass, param.CurrentDB, param.SQL, ast.GetMeta(queryAudit.Stmt, nil).Tables(""), param.Domains, traceID)
	if err != nil {
		log.Errorf("fetch table index err, db:%s,error:%s, trace_id:%s", param.CurrentDB, err.Error(), traceID)
		return
	}

	// +++++++++++++++++++++启发式规则建议[开始]+++++++++++++++++++++++{
	log.Infof("start of heuristic advisor Query: %s, trace_id:%s", queryAudit.Query, traceID)
	heuristicSuggestStart := time.Now()
	heuristicSuggest = advisor.CheckHeuristicRules(queryAudit) //这个函数返回的是违法的规则列表
	sysMetrics.CollectServiceMetrics("Service.heuristicSuggest", sysMetrics.GetStatus(err), time.Since(heuristicSuggestStart))
	//这是一个Prometheus监控埋点函数，用来记录服务内部每个函数调用的性能指标（耗时、成功/失败状态） TODO但是这里的err不是我们启发式建议返回的 所以一直都是success
	// +++++++++++++++++++++启发式规则建议[结束]+++++++++++++++++++++++}

	//开始初始化索引优化建议器
	log.Infof("start of index advisor Query: %s, trace_id:%s", queryAudit.Query, traceID)
	idxAdvisor, err = advisor.NewAdvisor(param.CurrentDB, tableInfo, *queryAudit, traceID)
	//创建IndexAdvisor实例，会自动从SQL语句的抽象语法树(AST)中提取各种索引相关的条件信息，为后续的索引优化建议提供基础数据
	if err != nil {
		log.Errorf("NewAdvisor err, db:%s,error:%s, trace_id:%s", param.CurrentDB, err.Error(), traceID)
		return
	}

	// ++++++++++++++++依赖线上表数据的启发式建议[开始]++++++++++++++++++{
	//k是string v是rule
	//这里是返回了一个map[string]Rule 使用k,v 把key和value拿出来 赋值给之前的heuristicSuggest
	//之前的他装的是启发式建议不依赖线上表格 现在新增了 依赖线上表格 的启发式建议
	//TODO：可以拿其中的一两个规则看下怎么根据有线上表格以及没有线上表格怎么根据sql以及ast检查规则情况的
	for k, v := range idxAdvisor.HeuristicCheck() {
		heuristicSuggest[k] = v
	}
	// ++++++++++++++++依赖线上表数据的启发式建议[结束]++++++++++++++++++{

	// ++++++++++++++不能使用索引场景，过滤索引推荐[开始]+++++++++++++++++}
	//作用：检查当前违反的规则里面是否存在着“不能生成索引优化建议”的规则
	//比如你的sql在where里面使用了函数 你的sql使用了not in这种性能很低 会触发全表扫描 这没有索引优化建议的必要了 不再向下进行 直接return
	for item := range heuristicSuggest {
		if _, exist := advisor.UnableIndexScenes[item]; exist {
			indexSuggest = map[string]advisor.Rule{"CHECK.005": {
				Item:    "CHECK.005",
				Summary: "No suggestions，Please view the「Sql Audit」 tag on the left",
				Content: "Please optimize your sql statement according to 「Sql Audit」 first, and then we can provide you with index optimization suggestions.",
			}}
			return
		}
	}
	// ++++++++++++++不能使用索引场景，过滤索引推荐[结束]++++++++++++++++}

	// +++++++++++++++++++++索引优化建议[开始]+++++++++++++++++++++++++{
	indexSuggestStart := time.Now()
	indexSuggest = idxAdvisor.IndexAdvise().Format()
	sysMetrics.CollectServiceMetrics("Service.indexSuggest", sysMetrics.GetStatus(err), time.Since(indexSuggestStart))
	// +++++++++++++++++++++索引优化建议[结束]+++++++++++++++++++++++++}
	return
}

// 49行的参数：param.SQl param.CurrenDB,mysqlSuggest,indexSuggest,heuristicSuggest
func (s *Service) formatSuggest(sql, db string, suggests ...map[string]advisor.Rule) {
	//小写的是内部的 FormatSuggest是soar的包
	formatResult, _ := advisor.FormatSuggest(sql, db, "json", suggests...)
	for _, suggest := range suggests {
		for key := range suggest {
			if _, ok := formatResult[key]; !ok {
				delete(suggest, key)
			}
		}
	}
	return
}

// fetchIndexInfo 连接线上db，获取表、索引、列等信息
// 链接数据库的优先级：shadow影子库 数据分析库 从库 主库
func fetchTableInfo(user, pass, db, sql string, tables []string, domains []*cmdbService.Domain, traceID string) (indexes map[string]*database.TableInfo, err error) {
	var (
		s         *storeMysql.Session
		tableInfo *database.TableInfo
	)

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("optimize.fetchTableInfo", sysMetrics.GetStatus(err), time.Since(start))

	indexes = make(map[string]*database.TableInfo)
	for _, table := range tables {
		if table == "" { // 空表不要查询
			log.Infof("fetchTableInfo table name is empty to skip, db:%s, sql:%s, trace_id:%s", db, sql, traceID)
			continue
		}
		var atLeastSuccess bool
		// domians是一个数组 每个元素是主库从库影子库分析库这些 遍历每个元素 每个元素是一个域名
		for _, slaveDomain := range domains {
			// 这里创建一个链接数据库的凭证s 连接到从库 从库端口是slaveDomain.Port
			if s, err = storeMysql.NewSession(user, pass, db, slaveDomain.Domain, int32(slaveDomain.Port)); err != nil {
				log.Errorf("service fetchIndexInfo NewSession error:%s, trace_id:%s", err.Error(), traceID)
				continue
			}
			if tableInfo, err = s.ShowTableInfoWithTimeout(db, table, 5*time.Second, traceID); err != nil {
				log.Errorf("service fetchIndexInfo ShowTableIndexWithTimeout statement:%s ,session:%s,error:%s, trace_id:%s",
					table, s.String(), err.Error(), traceID)
				continue
			}
			_ = s.Close()
			if tableInfo != nil {
				atLeastSuccess = true
				indexes[table] = tableInfo
				break
			}
		}
		if !atLeastSuccess {
			return nil, fmt.Errorf("can`t fetch db.table:%s.%s index info", db, table)
		}
	}
	return indexes, nil
}

func isCrossDatabase(meta common.Meta, currentDB string) bool {
	// 解析语法树，获库表基本信息 获取以DB为key，基本信息为value的map
	//meta是一个数据库名字为key的map 长度大于1 涉及多个数据库
	if len(meta) > 1 {
		return true
	}
	if _, exist := meta[currentDB]; !exist && len(meta) == 1 {
		return true
	}
	return false
}

// 检验是不是非DML和DQL（select）如果是 返回ture 不是 返回false
func isNotDMLORDQL(audit *advisor.Query4Audit) (bool, string) {
	info := "Optimization recommendations that do not support %s"
	switch audit.Stmt.(type) {
	case *sqlparser.DDL:
		return true, fmt.Sprintf(info, "DDL")

	case *sqlparser.DBDDL:
		return true, fmt.Sprintf(info, "DBDDL")

	case *sqlparser.Use:
		return true, fmt.Sprintf(info, "Use")

	case *sqlparser.Show:
		return true, fmt.Sprintf(info, "Show")

	default:
		return false, ""
	}
}
