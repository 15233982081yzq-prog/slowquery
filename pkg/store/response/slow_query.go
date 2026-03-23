package response

import (
	mysqlParser "smart-slowquery/pkg/parser/mysql"

	"smart-slowquery/internal/util/hint"
	"smart-slowquery/pkg/store/request"

	"time"
)

// ------------------------------------------------------------- 接口查询结果 ----------------------------------------------------------------------------------------//

type SlowQuery struct {
	FingerID      string  `gorm:"column:finger_id"`
	FingerSql     string  `gorm:"column:finger_sql"`
	ClusterUUID   string  `gorm:"column:cluster_uuid"`
	DBName        string  `gorm:"column:database_name"`
	TotalTime     float64 `gorm:"column:total_time"`
	AvgTime       float64 `gorm:"column:avg_time"`
	Count         int     `gorm:"column:count"`
	NewAppearFlag int     `gorm:"new_appear_flag"`
}

func (sq *SlowQuery) GetAppearType() string {
	if sq.NewAppearFlag > 0 {
		return request.AppearTypeNewAppeared
	}
	return request.AppearTypeOriginal
}

type SlowQueryDBStatistic struct {
	FingerID       string  `gorm:"column:finger_id"`
	FingerSQL      string  `gorm:"column:finger_sql"`
	ClusterUUID    string  `gorm:"column:cluster_uuid"`
	DBName         string  `gorm:"column:database_name"`
	Count          int     `gorm:"column:count"`
	TotalQueryTime float64 `gorm:"column:total_query_time"`
	TotalLockTime  float64 `gorm:"column:total_lock_time"`
	AvgQueryTime   float64 `gorm:"column:avg_query_time"`
	Top80Time      float64 `gorm:"column:top80_time"`
	Role           string  `gorm:"column:role"`
}

type QueryStatement struct {
	Statement     string    `gorm:"column:slow_sql"`
	DbHost        string    `gorm:"column:instance_host"`
	DbPort        int32     `gorm:"column:instance_port"`
	QueryTime     float64   `gorm:"column:query_time"`
	LockTime      float64   `gorm:"column:lock_time"`
	NewAppearFlag int       `gorm:"column:is_new_appeared"`
	LogTime       time.Time `gorm:"column:log_time"`
	StatementType string
}

func (qs *QueryStatement) GetStatementRemoveTraceHint() string {
	return hint.RemoveHint(qs.Statement)
}

func (qs *QueryStatement) GetStatementType() (typ string) {
	typ, _ = mysqlParser.ParseSqlStatement(qs.GetStatementRemoveTraceHint())
	qs.StatementType = typ
	return typ
}

type ClientHostStats struct {
	ClientHost string `gorm:"column:client_host"`
	Count      int    `gorm:"column:count"`
}

type ClientUserStats struct {
	ClientUser string `gorm:"column:client_user"`
	Count      int    `gorm:"column:count"`
}

//--------------------------------------------------------------------- TOPIC N 统计结果 ----------------------------------------------------------------------------//

type CKDBQueryRank struct {
	Rank    []*CKDBQueryTime
	OrderBy string
	Time    string
	Env     string
}

type CKDBQueryTime struct {
	ClusterUUID string  `gorm:"cluster_uuid"`
	DBName      string  `gorm:"column:database_name"`
	Total       float64 `gorm:"column:total"`
	Count       int     `gorm:"column:count"`
}

type CKFingerQueryRank struct {
	Rank    []*CKFingerQueryTime
	OrderBy string
	Time    string
	Env     string
}

type CKFingerQueryTime struct {
	FingerID    string  `gorm:"column:finger_id"`
	FingerSql   string  `gorm:"column:finger_sql"`
	ClusterUUID string  `gorm:"column:cluster_uuid"`
	DBName      string  `gorm:"column:database_name"`
	Total       float64 `gorm:"column:total"`
	Count       int     `gorm:"column:count"`
}

// --------------------------------------------------------------------- new sql report ----------------------------------------------------------------------------//

type NewFingerReportRecord struct {
	DatabaseName     string    `gorm:"column:database_name"`
	ClusterUUID      string    `gorm:"column:cluster_uuid"`
	QueryFingerCount int       `gorm:"column:query_finger_count"`
	QuerySQLCount    int       `gorm:"column:query_sql_count"`
	Datetime         time.Time `gorm:"column:datetime"`
	Total            float64   `gorm:"column:total"`
}

// --------------------------------------------------------------------- finger_id with earliest_appear_time ----------------------------------------------------------------------------//

type FingerIDWithEarliestAppearTime struct {
	FingerID string    `gorm:"column:finger_id"`
	Earliest time.Time `gorm:"column:earliest_appear_time"`
}
