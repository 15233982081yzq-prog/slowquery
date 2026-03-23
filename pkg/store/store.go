package store

import (
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"
)

type CKStore interface {
	GetQueryStatementOne(req *request.SlowQueryStatement) (*response.QueryStatement, error)
	GetQueryStatements(req *request.SlowQueryStatementWithOrderBy) (*[]response.QueryStatement, error)
	GetQueryStatementsCount(req *request.SlowQueryStatementWithOrderBy) (int, error)
	GetClientUsers(req *request.SlowQueryClientUsers) ([]string, error)
	GetClientHostsStats(req *request.SlowQueryClientHostsStats) (*[]response.ClientHostStats, error)
	GetInstanceHosts(req *request.SlowQueryInstanceHosts) ([]string, error)
	GetSlowQueryDBStatistics(req *request.SlowQueryDBStatistic) (*[]response.SlowQueryDBStatistic, error)
	GetSlowQueryList(req *request.SlowQueryList) (*[]response.SlowQuery, error)
	GetLast7Days(req []string) ([]string, error)
	GetSlowQueryCount(req *request.SlowQueryCount) (int, error)
	GetDBSlowQueryRank(req *request.SlowQueryRank) (*response.CKDBQueryRank, error)
	GetFingerSlowQueryRank(req *request.SlowQueryRank) (*response.CKFingerQueryRank, error)
	GetNewFingerSlowQueryReportRecord(req *request.SlowQueryNewFinger) ([]*response.NewFingerReportRecord, error)
	CreateAlertOperatorLog(req *request.AlertOperatorLog) error
	CreateAlertMessage(req *request.AlertMessage) error
	UpdateAlertMessage(req *request.AlertMessage) error
	CreateAlertMute(req *request.AlertMute) error
	GetAlertMute(alertID string) ([]*response.AlertMute, error)
	UpdateMuteStatus(monitorMuteID, status string) error
	BatchUpdateMuteStatusToTTL(srcStatus, dstStatus string, ts int64) error
	GetAlertMessageByAlertID(alertID int64) (*response.AlertMessage, error)
	GetAlertMessageByMonitorRuleIdAndStatus(ruleID, status string) (message []*response.AlertMessage, err error)
	GetAlertMessageByAlertIDs(alertIDs []string) ([]*response.AlertMessage, error)
	GetAlertMessage(req *request.AlertMessageSearch) ([]*response.AlertMessageWithMuteOperator, error)
	GetAlertMessageCount(req *request.AlertMessageSearch) (int, error)
	GetAlertMessageCountBySpecCond(req *request.AlertMessageCountSearch) (int, error)
	UpdateAlertMessageStatus(alertId, status string) (err error)
	GetUnMutedAndTTLStatusMessageList() (resp []*response.AlertMessage, err error)
	DeleteMute(alertId string) error
}

type CKWriter interface {
	Append(slowQuery *request.SlowQueryLog) (err error, flushed bool)
	FlushAll() (err error)
}

type Session interface {
	Explain(sql string) (result []response.ExplainInfo, err error)
	String() string
	Close() error
}
