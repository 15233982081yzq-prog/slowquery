package request

import (
	timeUtil "smart-slowquery/internal/util/time"

	"smart-slowquery/internal/model/filebeat"

	"fmt"
	"time"
)

const (
	TimeFormat            = "2006-01-02 15:04:05"
	AppearTypeAll         = "all"
	AppearTypeNewAppeared = "new_appeared"
	AppearTypeOriginal    = "original"
)

type SlowQueryDistinctFinger struct {
	Offset    int
	Limit     int
	StartTime string
	EndTime   string
}

func (req *SlowQueryDistinctFinger) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryStatement struct {
	FingerID   string
	DbName     string
	DbEnv      string
	Instances  []string
	StartTime  int64
	EndTime    int64
	AppearType *AppearType
}

func BuildSlowQueryStatement(dbName, dbEnv, fingerID string, instances []string, startTime, endTime int64, appearType *AppearType) *SlowQueryStatement {
	return &SlowQueryStatement{
		FingerID:   fingerID,
		DbName:     dbName,
		DbEnv:      dbEnv,
		Instances:  instances,
		StartTime:  startTime,
		EndTime:    endTime,
		AppearType: appearType,
	}
}

func (req *SlowQueryStatement) Valid() error {
	if len(req.DbName) == 0 {
		return fmt.Errorf("SlowQueryStatement param failed, dbName is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("SlowQueryStatement param failed, dbEnv is empty")
	}
	if len(req.FingerID) == 0 {
		return fmt.Errorf("SlowQueryStatement param failed, fringerID is empty")
	}
	if req.AppearType == nil {
		return fmt.Errorf("SlowQueryStatement param failed, AppearType is empty")
	}
	return nil
}

func (req *SlowQueryStatement) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryStatementWithOrderBy struct {
	SlowQueryStatement
	OrderBy string
	Offset  int
	Limit   int
}

func BuildSlowQueryStatementWithOrderBy(dbName, dbEnv, fingerID, orderBy string, instances []string, startTime, endTime int64, offset, limit int) *SlowQueryStatementWithOrderBy {
	stmt := &SlowQueryStatementWithOrderBy{}
	stmt.FingerID = fingerID
	stmt.DbName = dbName
	stmt.DbEnv = dbEnv
	stmt.OrderBy = orderBy
	stmt.Instances = instances
	stmt.StartTime = startTime
	stmt.EndTime = endTime
	stmt.Offset = offset
	stmt.Limit = limit

	return stmt
}

func (req *SlowQueryStatementWithOrderBy) Valid() error {
	if len(req.DbName) == 0 {
		return fmt.Errorf("SlowQueryStatementWithOrderBy param failed, dbName is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("SlowQueryStatementWithOrderBy param failed, dbEnv is empty")
	}
	if len(req.FingerID) == 0 {
		return fmt.Errorf("SlowQueryStatementWithOrderBy param failed, fringerID is empty")
	}
	if len(req.OrderBy) == 0 {
		return fmt.Errorf("SlowQueryStatementWithOrderBy param failed, orderby is empty")
	}
	//if req.Limit <= 0 || req.Limit > 30 {
	//	return fmt.Errorf("SlowQueryStatementWithOrderBy param failed, limit:%d invalid", req.Limit)
	//}
	return nil
}

func (req *SlowQueryStatementWithOrderBy) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryClientUsers struct {
	FingerID  string
	DbName    string
	DbEnv     string
	Instances []string
	StartTime int64
	EndTime   int64
}

type SlowQueryDISTINCTFingerSQL struct {
	GroupBy   string
	StartTime time.Time
	EndTime   time.Time
}

type SlowQueryHitFlag struct {
	FingerSQL    string    `gorm:"column:finger_sql"`
	SlowSQL      string    `gorm:"column:slow_sql"`
	DatabaseName string    `gorm:"column:database_name"`
	ClusterUUID  string    `gorm:"column:cluster_uuid"`
	CreateTime   time.Time `gorm:"column:create_time"`
	IsHit        int       `gorm:"column:is_hit"`
}

func (req *SlowQueryHitFlag) TableName() string {
	return "slow_query_log_hit_rand"
}

func (req *SlowQueryDISTINCTFingerSQL) GetTableName() string {
	return "slow_query_log_all_rand"
}

func BuildSlowQueryClientUsers(dbName, dbEnv, fingerID string, instances []string, startTime, endTime int64) *SlowQueryClientUsers {
	return &SlowQueryClientUsers{
		FingerID:  fingerID,
		DbName:    dbName,
		DbEnv:     dbEnv,
		Instances: instances,
		StartTime: startTime,
		EndTime:   endTime,
	}
}

func (req *SlowQueryClientUsers) Valid() error {
	if len(req.DbName) == 0 {
		return fmt.Errorf("SlowQueryClientUsers param failed, dbName is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("SlowQueryClientUsers param failed, dbEnv is empty")
	}
	if len(req.FingerID) == 0 {
		return fmt.Errorf("SlowQueryClientUsers param failed, fringerID is empty")
	}
	return nil
}

func (req *SlowQueryClientUsers) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryClientHostsStats struct {
	DbName    string
	DbEnv     string
	FingerID  string
	Instances []string
	StartTime int64
	EndTime   int64
}

func BuildSlowQueryClientHostsStats(dbName, dbEnv, fingerID string, instances []string, startTime, endTime int64) *SlowQueryClientHostsStats {
	return &SlowQueryClientHostsStats{
		DbName:    dbName,
		DbEnv:     dbEnv,
		FingerID:  fingerID,
		Instances: instances,
		StartTime: startTime,
		EndTime:   endTime,
	}
}

func (req *SlowQueryClientHostsStats) Valid() error {
	if len(req.DbName) == 0 {
		return fmt.Errorf("GetInstanceHosts param failed, dbName is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("GetInstanceHosts param failed, dbEnv is empty")
	}
	return nil
}

func (req *SlowQueryClientHostsStats) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryInstanceHosts struct {
	DbName    string
	DbEnv     string
	StartTime int64
	EndTime   int64
}

func BuildSlowQueryInstanceHosts(dbName, dbEnv string, startTime, endTime int64) *SlowQueryInstanceHosts {
	return &SlowQueryInstanceHosts{
		DbName:    dbName,
		DbEnv:     dbEnv,
		StartTime: startTime,
		EndTime:   endTime,
	}
}

func (req *SlowQueryInstanceHosts) Valid() error {
	if len(req.DbName) == 0 {
		return fmt.Errorf("GetInstanceHosts param failed, dbName is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("GetInstanceHosts param failed, dbEnv is empty")
	}
	return nil
}

func (req *SlowQueryInstanceHosts) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryCount struct {
	DbNames      []string
	DbEnv        string
	ClusterUUIDs []string
	Instances    []string
	StartTime    int64
	EndTime      int64
	AppearType   *AppearType
}

func BuildSlowQueryCount(dbEnv string, clusterUUIDs, dbNames, instances []string, startTime, endTime int64, appearType *AppearType) *SlowQueryCount {
	return &SlowQueryCount{
		DbNames:      dbNames,
		DbEnv:        dbEnv,
		ClusterUUIDs: clusterUUIDs,
		Instances:    instances,
		StartTime:    startTime,
		EndTime:      endTime,
		AppearType:   appearType,
	}
}

func (req *SlowQueryCount) Valid() error {
	if len(req.DbNames) == 0 {
		return fmt.Errorf("SlowQueryCount param failed, dbNames is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("SlowQueryCount param failed, dbEnv is empty")
	}
	if len(req.ClusterUUIDs) == 0 {
		return fmt.Errorf("SlowQueryCount param failed, ClusterUUIDs is empty")
	}
	if req.AppearType == nil {
		return fmt.Errorf("SlowQueryCount param failed, AppearType is empty")
	}
	return nil
}

func (req *SlowQueryCount) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryList struct {
	DbEnv        string
	ClusterUUIDs []string
	DbNames      []string
	Instances    []string
	OrderBy      string
	Limit        int
	Offset       int
	StartTime    int64
	EndTime      int64
	AppearType   *AppearType
}

func BuildSlowQueryList(dbEnv, orderBy string, clusterUUIDs, dbNames, instances []string, limit, offset int, startTime, endTime int64, appearType *AppearType) *SlowQueryList {
	return &SlowQueryList{
		DbEnv:        dbEnv,
		ClusterUUIDs: clusterUUIDs,
		OrderBy:      orderBy,
		DbNames:      dbNames,
		Instances:    instances,
		Limit:        limit,
		Offset:       offset,
		StartTime:    startTime,
		EndTime:      endTime,
		AppearType:   appearType,
	}
}

func (req *SlowQueryList) Valid() error {
	if len(req.DbNames) == 0 {
		return fmt.Errorf("SlowQueryList param failed, DbNames is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("SlowQueryList param failed, dbEnv is empty")
	}
	if len(req.OrderBy) == 0 {
		return fmt.Errorf("SlowQueryList param failed, roler is empty")
	}
	if len(req.ClusterUUIDs) == 0 {
		return fmt.Errorf("SlowQueryList param failed, ClusterUUIDs is empty")
	}
	if req.AppearType == nil {
		return fmt.Errorf("SlowQueryCount param failed, AppearType is empty")
	}
	return nil
}

func (req *SlowQueryList) GetTableName() string {
	return "slow_query_log_all_rand"
}

type SlowQueryLog struct {
	FingerID      string  `gorm:"column:finger_id"`
	FingerSql     string  `gorm:"column:finger_sql"`
	Query         string  `gorm:"column:slow_sql"`
	Hint          string  `gorm:"column:hint"`
	ClusterUUID   string  `gorm:"column:cluster_uuid"`
	Host          string  `gorm:"column:instance_host"`
	Port          uint32  `gorm:"column:instance_port"`
	DBName        string  `gorm:"column:database_name"`
	DBEnv         string  `gorm:"column:environment"`
	QueryTime     float64 `gorm:"column:query_time"`
	LockTime      float64 `gorm:"column:lock_time"`
	ExaminedRows  uint64  `gorm:"column:examine_rows"`
	NumRows       uint64  `gorm:"column:num_rows"`
	AffectRows    uint64  `gorm:"column:affect_rows"`
	BytesSent     uint64  `gorm:"column:bytes_sent"`
	ClientIP      string  `gorm:"column:client_host"`
	ConnectionID  uint64  `gorm:"column:connection_id"`
	DefaultUser   string  `gorm:"column:default_user"`
	User          string  `gorm:"column:client_user"`
	LogTime       string  `gorm:"column:log_time;type:DateTime"`
	CreateTime    string  `gorm:"column:create_time"`
	Killed        uint8   `gorm:"column:killed"`
	LastErrno     int32   `gorm:"column:lastErrno"`
	IsNewAppeared int     `gorm:"column:is_new_appeared"`
	IndexSwitch   int     `gorm:"column:index_switch"`
	IsHitIndex    int     `gorm:"column:is_hit_index"`
	L1L2          string  `gorm:"column:l1l2"`
	Team          string  `gorm:"column:team"`
	Role          string  `gorm:"column:role"`
}

func (log *SlowQueryLog) TableName() string {
	return "slow_query_log_all_rand"
}

func BuildSlowQueryLog(slowQuery *filebeat.SlowQuery) *SlowQueryLog {
	//判断hint字段是否为空，需要主动填充"" 信息，提升ck的读取性能
	if len(slowQuery.SlowLog.Hint) == 0 {
		slowQuery.SlowLog.Hint = ""
	}

	return &SlowQueryLog{
		FingerID:      slowQuery.SlowLog.FingerID,
		FingerSql:     slowQuery.SlowLog.FingerSql,
		Query:         slowQuery.SlowLog.Query,
		Hint:          slowQuery.SlowLog.Hint,
		ClusterUUID:   slowQuery.Fields.ClusterUUID,
		Host:          slowQuery.Fields.InstanceHost,
		Port:          uint32(slowQuery.Fields.InstancePort),
		DBEnv:         slowQuery.Fields.Env,
		DBName:        slowQuery.SlowLog.CurrentDB,
		QueryTime:     slowQuery.SlowLog.QueryTime,
		LockTime:      slowQuery.SlowLog.LockTime,
		ExaminedRows:  uint64(slowQuery.SlowLog.ExaminedRows),
		NumRows:       uint64(slowQuery.SlowLog.NumRows),
		AffectRows:    uint64(slowQuery.SlowLog.AffectRows),
		BytesSent:     uint64(slowQuery.SlowLog.BytesSent),
		ClientIP:      slowQuery.SlowLog.ClientIP,
		ConnectionID:  uint64(slowQuery.SlowLog.ConnectionID),
		DefaultUser:   slowQuery.SlowLog.DefaultUser,
		User:          slowQuery.SlowLog.User,
		LogTime:       timeUtil.UnixTimeFormat(slowQuery.SlowLog.TimeStamp, timeUtil.SecondFormat),
		CreateTime:    timeUtil.UnixTimeFormat(time.Now().Unix(), timeUtil.SecondFormat),
		Killed:        uint8(slowQuery.SlowLog.Killed),
		LastErrno:     int32(slowQuery.SlowLog.LastErrno),
		IsNewAppeared: slowQuery.IsNewAppeared,
		IndexSwitch:   slowQuery.HasOpenIndexSwitch,
		IsHitIndex:    slowQuery.IsHitIndex,
		L1L2:          slowQuery.L1L2,
		Team:          slowQuery.Team,
		Role:          slowQuery.Role,
	}
}

type SlowQueryDBStatistic struct {
	DbEnv        string
	ClusterUUids []string
	DbNames      []string
	StartTime    int64
	EndTime      int64
}

func BuildSlowQueryDBStatistic(dbEnv string, clusterUUids, dbNames []string, startTime, endTime int64) *SlowQueryDBStatistic {
	return &SlowQueryDBStatistic{
		DbEnv:        dbEnv,
		ClusterUUids: clusterUUids,
		DbNames:      dbNames,
		StartTime:    startTime,
		EndTime:      endTime,
	}
}

func (req *SlowQueryDBStatistic) Valid() error {
	if len(req.ClusterUUids) == 0 {
		return fmt.Errorf("SlowQueryDBStatistic param failed, clusterUUids is empty")
	}
	if len(req.DbNames) == 0 {
		return fmt.Errorf("SlowQueryDBStatistic param failed, DbNames is empty")
	}
	if len(req.DbEnv) == 0 {
		return fmt.Errorf("SlowQueryDBStatistic param failed, dbEnv is empty")
	}
	if req.EndTime <= req.StartTime {
		return fmt.Errorf("SlowQueryDBStatistic param failed, startTime >= endTime")
	}
	return nil
}

func (req *SlowQueryDBStatistic) GetTableName() string {
	return "slow_query_log_all_rand"
}

type AppearType struct {
	name string
	sign int
}

func BuildAppearType(name string, sign int) *AppearType {
	return &AppearType{name: name, sign: sign}
}

func (nas *AppearType) IsAll() bool {
	return AppearTypeAll == nas.name
}

func (nas *AppearType) IsOriginal() bool {
	return AppearTypeOriginal == nas.name
}

func (nas *AppearType) IsNewAppeared() bool {
	return AppearTypeNewAppeared == nas.name
}

func (nas *AppearType) GetName() string {
	return nas.name
}

func (nas *AppearType) GetSign() int {
	return nas.sign
}
