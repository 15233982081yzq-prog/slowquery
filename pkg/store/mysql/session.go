package mysql

import (
	"context"
	dbSql "database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sysMetrics "smart-slowquery/pkg/metrics/platform"
	"smart-slowquery/thrid-party/soar-dev/common"
	"smart-slowquery/thrid-party/soar-dev/database"

	"smart-slowquery/internal/util/function"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/response"

	"github.com/jinzhu/gorm"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var CtxTimeoutErr = fmt.Errorf("context timeout error")

type Session struct {
	dbUser   string
	dbPwd    string
	db       *gorm.DB
	host     string
	port     int32
	dbName   string
	command  string
	maxRetry int
}

func NewSession(user, pwd, dbName, host string, port int32) (*Session, error) {
	s := &Session{
		dbUser:   user,
		dbPwd:    pwd,
		host:     host,
		dbName:   dbName,
		port:     port,
		maxRetry: 3,
	}
	return s, s.initDB()
}

func (s *Session) initDB() (err error) {
	if s.db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=10s", s.dbUser, s.dbPwd, s.host, s.port, s.dbName)); err != nil {
		log.Errorf("NewSession initDB %s,error:%s", s.String(), err.Error())
	}
	return err
}

func (s *Session) ExplainWithTimeout(sql string, timeout time.Duration) (list []*response.ExplainInfo, err error) {
	var (
		connID int //
		start  = time.Now()
	)

	defer func() {
		sysMetrics.CollectServiceMetrics("session.ExplainWithTimeout", sysMetrics.GetStatus(err), time.Since(start))
	}()

	log.Infof("ExplainWithTimeout sql:%s,timeout:%v", sql, timeout)
	if list, connID, err = s.explainWithTimeout(sql, timeout); //这里是开始执行解释
	err != nil && errors.Is(err, context.DeadlineExceeded) && connID > 0 {
		log.Errorf("session explain connID:%d ,sql:%s ,timeout:%v ,error:%s", connID, sql, timeout, err.Error())
		if killErr := s.killQueryTimeout(connID, timeout); killErr != nil { //TODO :异常多种多样，且执行kill时，被kill的对象连接也可能已经执行完毕了，如何判断kill失败呢？
			log.Errorf("session explain connID:%d ,sql:%s ,ctx_timeout kill error:%s", connID, sql, killErr.Error())
			return list, err
		} else {
			log.Infof("session explain connID:%d ,sql:%s ,ctx_timeout was killed", connID, sql)
			return list, err
		}
	}

	return list, err
}

func (s *Session) ShowTableInfoWithTimeout(db, table string, timeout time.Duration, traceID string) (indexInfo *database.TableInfo, err error) {
	var (
		connID int
		start  = time.Now()
	)

	defer func() {
		sysMetrics.CollectServiceMetrics("session.ShowTableIndexWithTimeout", sysMetrics.GetStatus(err), time.Since(start))
	}()
	// 获取索引结构，获取表-行信息
	log.Infof("ShowTableIndexWithTimeout table:%s,timeout:%v, trace_id:%s", table, timeout, traceID)
	if indexInfo, connID, err = s.showTableInfoWithTimeout(db, table, timeout); err != nil && errors.Is(err, context.DeadlineExceeded) && connID > 0 {
		log.Errorf("session show index connID:%d ,table:%s ,timeout:%v ,error:%s, trace_id:%s", connID, table, timeout, err.Error(), traceID)
		if killErr := s.killQueryTimeout(connID, timeout); killErr != nil {
			log.Errorf("session show index connID:%d ,table:%s ,ctx_timeout kill error:%s, trace_id:%s", connID, table, killErr.Error(), traceID)
			return indexInfo, err
		} else {
			log.Infof("session show index connID:%d ,table:%s ,ctx_timeout was killed, trace_id:%s", connID, table, traceID)
			return indexInfo, err
		}
	}

	return indexInfo, err
}

func (s *Session) GetWithLimit(limit int, currMeta interface{}, query interface{}) error {
	return s.db.Where(query).Limit(limit).Find(currMeta).Error
}

func (s *Session) Create(data interface{}) error {
	return s.db.Create(data).Error
}

func (s *Session) Save(data interface{}) error {
	return s.db.Save(data).Error
}

func (s *Session) String() string {
	return fmt.Sprintf("user:%s,host:%s,port:%d,dbName:%s", s.dbUser, s.host, s.port, s.dbName)
}

//------------------------------------ 内部函数 ---------------------------------------------------------//

func (s *Session) explainWithTimeout(sql string, timeout time.Duration) (list []*response.ExplainInfo, connID int, err error) {
	var (
		conn *dbSql.Conn // 数据库连接 每次都是拿这个去做查询 id 和 explain
		rows *dbSql.Rows // 查询结果集
	)
	// ===== 步骤1：获取数据库连接 =====
	if conn, err = s.db.DB().Conn(context.Background()); err != nil {
		log.Errorf("explainWithTimeout conn error:%s", err.Error())
		return nil, -1, err
	}
	defer conn.Close() // 确保数据库连接关闭

	// ===== 步骤2：查询当前连接ID（用于后续KILL） =====
	// 2.1 获取一个上下文计时器ctx 时间是五秒钟 调用WithTimeout Go会为当前的上下文设置一个定时器当你使用conn进行查询的时候传入这个ctx
	// 传入这个ctx之后（QueryContext）如果超过了五秒 这个函数就会返回一个context.DeadlineExceeded错误
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 2.2 开始查询connection_id
	// 两个ctx是不同的上下文 他们两个使用的是一个和数据库的连接 不同的查询使用不同的ctx计时
	// 超时的时候 QueryContext会检测，返回一个context.DeadlineExceeded，在调用本函数的外面会检测是否这个错误进行选择kill掉这个数据库连接
	if rows, err = conn.QueryContext(ctx, "select connection_id();"); err != nil {
		log.Errorf("explainWithTimeout get ConnectionID error:%s", err.Error())
		return nil, -1, err
	}
	// Go的特性 使用Gorm连接数据库获取链接conn之后每次都是使用这个conn去数据库做查询 他会自动把返回结果放到你的&里面
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&connID); err != nil { // 把connection_id放到connID
			log.Errorf("explainWithTimeout scan ConnectionID error:%s", err.Error())
			return nil, -1, fmt.Errorf("executeWithTimeout get connectionID error:%s", err.Error())
		}
	}
	// 现在已经拿到了cornid通过row.next赋值给了connID

	// ===== 步骤3：执行 EXPLAIN（带超时控制） =====
	// 3.1 获取一个计时器 时长5s ctx和ctx2分开是为了避免 上面拿到connection_id的时间 影响这里的EXPLAIN查询时间
	ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
	defer cancel2()
	// 执行：EXPLAIN {原始SQL}
	if rows, err = conn.QueryContext(ctx2, fmt.Sprintf("EXPLAIN %s", sql)); err != nil {
		log.Errorf("explainWithTimeout explain sql:%s, error:%s", sql, err.Error())
		return nil, connID, err
	}
	// ===== 步骤4：解析 EXPLAIN 结果 ===== 自动把结果放在你的&里
	for rows.Next() {
		var explainInfo response.ExplainInfo
		if err = rows.Scan(&explainInfo.ID,
			&explainInfo.SelectType,
			&explainInfo.Table,
			&explainInfo.Partitions,
			&explainInfo.Type,
			&explainInfo.PossibleKeys,
			&explainInfo.Key,
			&explainInfo.KeyLen,
			&explainInfo.Ref,
			&explainInfo.Rows,
			&explainInfo.Filtered,
			&explainInfo.Extra); err != nil {
			return nil, connID, err
		}
		list = append(list, &explainInfo) //list没有在函数内部显示声明 在函数名里 所以最后只需一个return 就可以自动返回list
	}

	return
}

func (s *Session) showTableInfoWithTimeout(db, table string, timeout time.Duration) (tableInfo *database.TableInfo, connID int, err error) {
	var (
		conn      *dbSql.Conn
		rows      *dbSql.Rows
		collation string
	)

	tableInfo = new(database.TableInfo)

	if conn, err = s.db.DB().Conn(context.Background()); err != nil {
		log.Errorf("show index conn error:%s", err.Error())
		return nil, -1, err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 依旧先查询连接的id
	if rows, err = conn.QueryContext(ctx, "select connection_id();"); err != nil {
		log.Errorf("showIndexWithTimeout get ConnectionID error:%s", err.Error())
		return nil, -1, err
	}

	defer func() { _ = conn.Close() }()
	for rows.Next() {
		if err = rows.Scan(&connID); err != nil {
			log.Errorf("showIndexWithTimeout scan ConnectionID error:%s", err.Error())
			return nil, -1, fmt.Errorf("showIndexWithTimeout get connectionID error:%s", err.Error())
		}
	}

	// 获取表索引结构
	if tableInfo.IndexInfo, err = s.showTableIndexWithTimeout(db, table, timeout); err != nil {
		return nil, connID, err
	}

	// 获取表-列索引结构
	if tableInfo.ColInfo, err = s.showTableColWithTimeout(db, table, timeout); err != nil {
		return nil, connID, err
	}

	// 获取表-编码索引结构
	if collation, err = s.showTableCollationWithTimeout(db, table, timeout); err != nil {
		return nil, connID, err
	}

	// 获取库-表版本信息
	if tableInfo.Version, err = s.showDBVersionWithTimeout(db, table, timeout); err != nil {
		return nil, connID, err
	}

	for _, v := range tableInfo.ColInfo {
		if v.Collation == "" && collation != "" {
			v.Collation = collation
			if split := strings.Split(collation, "_"); len(split) > 1 {
				v.Character = split[0]
			}
		}
	}
	return tableInfo, connID, nil
}

func (s *Session) showTableIndexWithTimeout(db, table string, timeout time.Duration) (indexInfo *database.TableIndexInfo, err error) {
	var (
		conn *dbSql.Conn
		rows *dbSql.Rows
	)

	indexInfo = database.NewTableIndexInfo(table)

	if conn, err = s.db.DB().Conn(context.Background()); err != nil {
		log.Errorf("show index conn error:%s", err.Error())
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 获取表索引结构
	if rows, err = conn.QueryContext(ctx, fmt.Sprintf("show index from `%s`.`%s`", db, table)); err != nil {
		log.Errorf("showIndexWithTimeout  db:%s, table:%s, error:%s", db, table, err.Error())
		return nil, err
	}

	ti := database.TableIndexRow{}
	indexFields := make([]interface{}, 0)
	fields := map[string]interface{}{
		"Table":         &ti.Table,
		"Non_unique":    &ti.NonUnique,
		"Key_name":      &ti.KeyName,
		"Seq_in_index":  &ti.SeqInIndex,
		"Column_name":   &ti.ColumnName,
		"Collation":     &ti.Collation,
		"Cardinality":   &ti.Cardinality,
		"Sub_part":      &ti.SubPart,
		"Packed":        &ti.Packed,
		"Null":          &ti.Null,
		"Index_type":    &ti.IndexType,
		"Comment":       &ti.Comment,
		"Index_comment": &ti.IndexComment,
		"Visible":       &ti.Visible,
		"Expression":    &ti.Expression,
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var colByPass []byte
	for _, col := range cols {
		if _, ok := fields[col]; ok {
			indexFields = append(indexFields, fields[col])
		} else {
			indexFields = append(indexFields, &colByPass)
		}
	}
	for rows.Next() {
		if err = rows.Scan(indexFields...); err != nil {
			return nil, err
		}
		indexInfo.Rows = append(indexInfo.Rows, ti)
	}
	return
}

func (s *Session) showTableColWithTimeout(db, table string, timeout time.Duration) (colInfo []*common.Column, err error) {
	var (
		conn *dbSql.Conn
		rows *dbSql.Rows
	)

	colInfo = make([]*common.Column, 0)

	if conn, err = s.db.DB().Conn(context.Background()); err != nil {
		log.Errorf("show index conn error:%s", err.Error())
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 获取表-行结构
	if rows, err = conn.QueryContext(ctx, fmt.Sprintf("SELECT "+
		"c.COLUMN_NAME,c.TABLE_NAME,c.TABLE_SCHEMA,c.COLUMN_TYPE,c.CHARACTER_SET_NAME, c.COLLATION_NAME "+
		"FROM `INFORMATION_SCHEMA`.`COLUMNS` as c where c.table_schema = '%s' and c.table_name='%s'", db, table)); err != nil {
		log.Errorf("showTableInfoWithTimeout  db:%s, table:%s, error:%s", db, table, err.Error())
		return nil, err
	}

	for rows.Next() {
		var character, collation []byte
		col := common.Column{}
		if err = rows.Scan(&col.Name, &col.Table, &col.DB, &col.DataType, &character, &collation); err != nil {
			return nil, err
		}

		col.Character = string(character)
		col.Collation = string(collation)

		colInfo = append(colInfo, &col)
	}
	return colInfo, nil
}

func (s *Session) showTableCollationWithTimeout(db, table string, timeout time.Duration) (collation string, err error) {
	var (
		conn *dbSql.Conn
		rows *dbSql.Rows
	)

	if conn, err = s.db.DB().Conn(context.Background()); err != nil {
		log.Errorf("show index conn error:%s", err.Error())
		return "", err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 获取表-行结构
	if rows, err = conn.QueryContext(ctx, fmt.Sprintf("SELECT `t`.`TABLE_COLLATION` FROM `INFORMATION_SCHEMA`.`TABLES` AS `t` "+
		"WHERE `t`.`TABLE_NAME`='%s' AND `t`.`TABLE_SCHEMA` = '%s'", table, db)); err != nil {
		log.Errorf("showTableInfoWithTimeout  db:%s, table:%s, error:%s", db, table, err.Error())
		return "", err
	}

	var tbCollation []byte
	if rows.Next() {
		if err = rows.Scan(&tbCollation); err != nil {
			return "", err
		}
	}
	_ = rows.Close()

	return string(tbCollation), nil
}

func (s *Session) showDBVersionWithTimeout(db, table string, timeout time.Duration) (version int, err error) {
	var (
		conn *dbSql.Conn
		rows *dbSql.Rows
	)

	version = 99999

	if conn, err = s.db.DB().Conn(context.Background()); err != nil {
		log.Errorf("show index conn error:%s", err.Error())
		return version, err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 获取表-行结构
	if rows, err = conn.QueryContext(ctx, "select @@version"); err != nil {
		log.Errorf("showTableInfoWithTimeout  db:%s, table:%s, error:%s", db, table, err.Error())
		return version, err
	}

	// MariaDB https://mariadb.com/kb/en/library/comment-syntax/
	// MySQL https://dev.mysql.com/doc/refman/8.0/en/comments.html
	var versionStr string
	var versionSeg []string

	if rows.Next() {
		if err = rows.Scan(&versionStr); err != nil {
			return version, err
		}
	}
	_ = rows.Close()

	versionStr = strings.Split(versionStr, "-")[0]
	versionSeg = strings.Split(versionStr, ".")
	if len(versionSeg) == 3 {
		versionStr = fmt.Sprintf("%s%02s%02s", versionSeg[0], versionSeg[1], versionSeg[2])
		version, err = strconv.Atoi(versionStr)
	}

	return version, err
}

func (s *Session) killQueryTimeout(connID int, timeout time.Duration) (err error) {
	var killerConn *dbSql.Conn

	killerCtx, killerCancel := context.WithTimeout(context.Background(), timeout)
	defer killerCancel()

	if killerConn, err = s.db.DB().Conn(killerCtx); err != nil {
		return err
	}
	defer killerConn.Close()

	_, err = killerConn.QueryContext(killerCtx, fmt.Sprintf("kill %d", connID))
	return err
}

func (s *Session) rawScan(sqlStr string, dest interface{}) (err error) {
	//执行explain，自动重试
	return function.Retry("session rawScan", func() error {
		if err = s.db.Raw(sqlStr).Scan(dest).Error; err == nil {
			return nil
		}

		if err != mysqlDriver.ErrInvalidConn {
			log.Errorf("session:%s ,error:%s", s.String(), err.Error())
			return err
		}

		return s.initConnection()
	}, s.maxRetry)
}

func (s *Session) initConnection() (err error) {
	// 连接断开无效时,自动重试
	return function.Retry("initConnection", func() error {
		if s.dbName == "" {
			err = s.db.DB().Ping()
		} else {
			err = s.db.Exec(fmt.Sprintf("USE `%s`", s.dbName)).Error
		}
		return err
	}, s.maxRetry)
}

func (s *Session) Close() error {
	log.Infof("session closed")
	return s.db.DB().Close()
}
