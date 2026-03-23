package clickhouse

import (
	errStore "smart-slowquery/internal/util/errors"
	timeUtil "smart-slowquery/internal/util/time"
	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"

	"fmt"
	"time"
)

type CKStore struct {
	client *Client
}

func NewCKStore(cliConf *conf.CKCli) (*CKStore, error) {
	if cliConf == nil {
		return nil, fmt.Errorf("ck_cliConf or ck_flusherConf param is empty")
	}
	var (
		ckCli *Client
		err   error
	)

	if ckCli, err = NewClient(cliConf); err != nil {
		return nil, err
	}

	return &CKStore{
		client: ckCli,
	}, nil
}

func (ck *CKStore) GetQueryStatementsCount(req *request.SlowQueryStatementWithOrderBy) (count int, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetQueryStatementsCount", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetQueryStatementsCount request:%v ,error:%s", req, err.Error())
		return -1, err
	}

	if count, err = ck.client.GetQueryStatementsCount(req); err != nil {
		log.Errorf("ck reader GetQueryStatementsCount request:%v ,error:%s", req, err.Error())
		return -1, err
	}
	return count, nil
}

func (ck *CKStore) GetQueryStatements(req *request.SlowQueryStatementWithOrderBy) (stmts *[]response.QueryStatement, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetQueryStatements", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetQueryStatements request:%v ,error:%s", req, err.Error())
		return nil, err
	}

	if stmts, err = ck.client.GetQueryStatements(req); err == nil && len(*stmts) == 0 {
		log.Warningf("ck reader GetQueryStatements request:%v ,not found query statement", req)
		return nil, errStore.NotFoundError
	}
	return stmts, err
}

func (ck *CKStore) GetQueryStatementOne(req *request.SlowQueryStatement) (stmt *response.QueryStatement, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetQueryStatementOne", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetQueryStatementOne request:%v ,error:%s", req, err.Error())
		return nil, err
	}

	if stmt, err = ck.client.GetFirstQueryStatement(req); err == nil && len(stmt.Statement) == 0 {
		log.Warningf("ck reader GetQueryStatementOne request:%v ,not found query statement", req)
		return nil, errStore.NotFoundError
	}
	return stmt, err
}

func (ck *CKStore) GetClientUsers(req *request.SlowQueryClientUsers) (users []string, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetClientUsers", sysMetrics.GetStatus(err), time.Since(start))
	}()

	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetClientUsers request:%v ,error:%s", req, err.Error())
		return nil, err
	}

	if users, err = ck.client.getClientUsers(req); err == nil && len(users) == 0 {
		log.Warningf("ck reader GetClientUsers request:%v ,not found query statement", req)
		return nil, errStore.NotFoundError
	}

	return users, err
}

func (ck *CKStore) GetClientHostsStats(req *request.SlowQueryClientHostsStats) (*[]response.ClientHostStats, error) {
	var err error

	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetClientHosts", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetClientHosts request:%v ,error:%s", req, err.Error())
		return nil, err
	}
	return ck.client.GetClientHostsStats(req)
}

func (ck *CKStore) GetInstanceHosts(req *request.SlowQueryInstanceHosts) ([]string, error) {
	var err error

	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetInstanceHosts", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetInstanceHosts request:%v ,error:%s", req, err.Error())
		return nil, err
	}

	return ck.client.GetInstanceHosts(req)
}

func (ck *CKStore) GetSlowQueryList(req *request.SlowQueryList) (*[]response.SlowQuery, error) {
	var err error

	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetSlowQueryList", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetSlowQueryList request:%v ,error:%s", req, err.Error())
		return nil, err
	}

	return ck.client.GetSlowQueryList(req)
}

func (ck *CKStore) GetLast7Days(req []string) ([]string, error) {
	return ck.client.GetLast7Days(req)
}

func (ck *CKStore) GetSlowQueryDBStatistics(req *request.SlowQueryDBStatistic) (*[]response.SlowQueryDBStatistic, error) {
	var err error

	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetSlowQueryDBStatistics", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetSlowQueryDBStatistics request:%v ,error:%s", req, err.Error())
		return nil, err
	}

	return ck.client.GetSlowQueryDBStatistic(req)
}

func (ck *CKStore) GetSlowQueryCount(req *request.SlowQueryCount) (total int, err error) {

	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("GetSlowQueryCount", sysMetrics.GetStatus(err), time.Since(start))
	}()
	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetSlowQueryList request:%v ,error:%s", req, err.Error())
		return -1, err
	}

	return ck.client.GetSlowQueryCount(req)
}

// --------------------------------------------------------------------- SLowQuery 报表 ----------------------------------------------------------------------------//

func (ck *CKStore) GetDBSlowQueryRank(req *request.SlowQueryRank) (*response.CKDBQueryRank, error) {
	var (
		err   error
		list  *[]response.CKDBQueryTime
		start = time.Now()
	)

	defer func() {
		sysMetrics.CollectStoreMetrics("GetDBSlowQueryRank", sysMetrics.GetStatus(err), time.Since(start))
	}()

	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetDBSlowQueryRank param invalid, order:%s ,top:%d ,start:%s ,end:%s", req.OrderBy, req.Top, req.StartTime.GoString(), req.EndTime.GoString())
		return nil, err
	}

	if list, err = ck.client.DBSlowQueryRank(req); err != nil {
		return nil, err
	}

	report := &response.CKDBQueryRank{}
	for i := range *list {
		report.Rank = append(report.Rank, &(*list)[i])
	}
	report.Time = timeUtil.UnixTimeFormat(req.StartTime.Unix(), timeUtil.DayFormat)
	report.OrderBy = req.OrderBy
	report.Env = req.DBEnv

	return report, nil
}

func (ck *CKStore) GetFingerSlowQueryRank(req *request.SlowQueryRank) (*response.CKFingerQueryRank, error) {
	var (
		err   error
		list  *[]response.CKFingerQueryTime
		start = time.Now()
	)

	defer func() {
		sysMetrics.CollectStoreMetrics("GetFingerSlowQueryRank", sysMetrics.GetStatus(err), time.Since(start))
	}()

	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetFingerSlowQueryRank param invalid, order:%s ,top:%d ,start:%s ,end:%s", req.OrderBy, req.Top, req.StartTime.GoString(), req.EndTime.GoString())
		return nil, err
	}

	if list, err = ck.client.FingerSlowQueryRank(req); err != nil {
		return nil, err
	}
	log.Infof("ck reader FingerSlowQueryRank rank size:%d", len(*list))

	report := &response.CKFingerQueryRank{}
	for i := range *list {
		report.Rank = append(report.Rank, &(*list)[i])
	}
	report.Time = timeUtil.UnixTimeFormat(req.StartTime.Unix(), timeUtil.DayFormat)
	report.OrderBy = req.OrderBy
	report.Env = req.DBEnv

	return report, nil
}

func (ck *CKStore) GetNewFingerSlowQueryReportRecord(req *request.SlowQueryNewFinger) ([]*response.NewFingerReportRecord, error) {
	var (
		err   error
		list  []*response.NewFingerReportRecord
		start = time.Now()
	)

	defer func() {
		sysMetrics.CollectStoreMetrics("GetNewFingerSlowQueryReportRecord", sysMetrics.GetStatus(err), time.Since(start))
	}()

	if list, err = ck.client.NewFingerSlowQueryReportRecord(req); err != nil {
		return nil, err
	}
	log.Infof("ck reader FingerSlowQueryRank rank size:%d", len(list))

	return list, nil
}

// ----------------------------------------------------------------------alert service ----------------------------------------------------------------------------//

func (ck *CKStore) CreateAlertOperatorLog(req *request.AlertOperatorLog) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("CreateAlertOperatorLog", sysMetrics.GetStatus(err), time.Since(start))

	if req == nil {
		log.Errorf("ck client SaveAlertOperatorLog param is nil")
		return fmt.Errorf("ckClient SaveAlertOperatorLog failed, param is nil")
	}

	if err = req.Valid(); err != nil {
		log.Errorf("ck store SaveAlertOperatorLog error:%s", err.Error())
		return err
	}

	err = ck.client.CreateAlertOperatorLog(req)
	return err
}

func (ck *CKStore) CreateAlertMessage(req *request.AlertMessage) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("CreateAlertMessage", sysMetrics.GetStatus(err), time.Since(start))

	if req == nil {
		log.Errorf("ck client CreateAlertMessage param is nil")
		return fmt.Errorf("ckClient CreateAlertMessage failed, param is nil")
	}

	if err = req.Valid(); err != nil {
		log.Errorf("ck store CreateAlertMessage error:%s", err.Error())
		return err
	}
	err = ck.client.CreateAlertMessage(req)
	return err
}

func (ck *CKStore) UpdateAlertMessage(req *request.AlertMessage) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("UpdateAlertMessage", sysMetrics.GetStatus(err), time.Since(start))

	if req == nil {
		log.Errorf("ck client UpdateAlertMessage param is nil")
		return fmt.Errorf("ckClient UpdateAlertMessage failed, param is nil")
	}

	if err = req.Valid(); err != nil {
		log.Errorf("ck store UpdateAlertMessage error:%s", err.Error())
		return err
	}
	err = ck.client.UpdateAlertMessage(req)
	return err
}

func (ck *CKStore) GetAlertMessageByAlertID(alertID int64) (message *response.AlertMessage, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMessageByAlertID", sysMetrics.GetStatus(err), time.Since(start))

	message, err = ck.client.GetAlertMessageByAlertID(alertID)
	return message, err
}

func (ck *CKStore) GetAlertMessageByMonitorRuleIdAndStatus(ruleID, status string) (messages []*response.AlertMessage, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMessageByMonitorRuleIdAndStatus", sysMetrics.GetStatus(err), time.Since(start))

	messages, err = ck.client.GetAlertMessageByMonitorRuleIdAndStatus(ruleID, status)
	return messages, err
}

func (ck *CKStore) GetAlertMessageByAlertIDs(alertIDs []string) (messages []*response.AlertMessage, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMessageByAlertIDs", sysMetrics.GetStatus(err), time.Since(start))

	messages, err = ck.client.GetAlertMessageByAlertIDs(alertIDs)
	return messages, err
}

func (ck *CKStore) GetAlertMessage(req *request.AlertMessageSearch) (messages []*response.AlertMessageWithMuteOperator, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMessage", sysMetrics.GetStatus(err), time.Since(start))

	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetAlertMessage param invalid, error:%s", err.Error())
		return nil, err
	}

	messages, err = ck.client.GetAlertMessage(req)
	return messages, err
}

func (ck *CKStore) GetAlertMessageCount(req *request.AlertMessageSearch) (count int, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMessageCount", sysMetrics.GetStatus(err), time.Since(start))

	if err = req.Valid(); err != nil {
		log.Errorf("ck reader GetAlertMessageCount param invalid, error:%s", err.Error())
		return 0, err
	}

	count, err = ck.client.GetAlertMessageCount(req)
	return count, err
}

func (ck *CKStore) GetUnMutedAndTTLStatusMessageList() (alertMessages []*response.AlertMessage, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetUnMutedAndTTLStatusMessageList", sysMetrics.GetStatus(err), time.Since(start))

	alertMessages, err = ck.client.GetUnMutedAndTTLStatusMessageList(&request.AlertNoMutedMessage{})
	return
}

func (ck *CKStore) DeleteMute(alertId string) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("DeleteMute", sysMetrics.GetStatus(err), time.Since(start))

	err = ck.client.DeleteAlertMute(&request.AlertMute{MonitorAlertID: alertId})
	return
}

func (ck *CKStore) GetAlertMessageCountBySpecCond(req *request.AlertMessageCountSearch) (count int, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMessageCountBySpecCond", sysMetrics.GetStatus(err), time.Since(start))

	count, err = ck.client.GetAlertMessageCountBySpecCond(req)
	return
}

func (ck *CKStore) UpdateAlertMessageStatus(alertId, status string) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("UpdateAlertMessageStatus", sysMetrics.GetStatus(err), time.Since(start))

	err = ck.client.UpdateAlertMessageStatus(&request.AlertMessage{
		MonitorAlertID: alertId,
		Status:         status,
		UpdateTime:     time.Now(),
	})
	return
}

func (ck *CKStore) CreateAlertMute(req *request.AlertMute) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("CreateAlertMute", sysMetrics.GetStatus(err), time.Since(start))

	if err = req.Valid(); err != nil {
		log.Errorf("ck reader CreateAlertMute param invalid, error:%s", err.Error())
		return err
	}

	err = ck.client.CreateAlertMute(req)
	return
}

func (ck *CKStore) GetAlertMute(alertID string) (alertMutes []*response.AlertMute, err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertMute", sysMetrics.GetStatus(err), time.Since(start))

	alertMutes, err = ck.client.GetAlertMuteByAlertID(alertID)
	return
}

func (ck *CKStore) UpdateMuteStatus(monitorMuteID, status string) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("UpdateMuteStatus", sysMetrics.GetStatus(err), time.Since(start))

	err = ck.client.UpdateMuteStatus(&request.AlertMute{
		MonitorMuteID: monitorMuteID,
		Status:        status,
		UpdateTime:    time.Now(),
	})
	return
}

func (ck *CKStore) BatchUpdateMuteStatusToTTL(srcStatus, dstStatus string, ts int64) (err error) {
	var start = time.Now()
	defer sysMetrics.CollectStoreMetrics("BatchUpdateMuteStatusToTTL", sysMetrics.GetStatus(err), time.Since(start))

	err = ck.client.BatchUpdateMuteStatusToTTL(&request.AlertMute{
		Status:     srcStatus,
		UpdateTime: time.Now(),
	}, dstStatus, ts)
	return
}
