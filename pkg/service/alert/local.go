package alert

import (
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	requestMeta "smart-slowquery/pkg/store/request"

	"smart-slowquery/internal/model/alert"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/store/response"

	"time"
)

// From local mysql db to query result

func (s *Service) FindAlertRuleFromDBByUUID(uuid string) (*alert.RuleTab, error) {
	return alert.FindAlertRuleByUUID(s.db, uuid)
}

func (s *Service) FindAlertChannelFromDBByUUID(uuid string) (*alert.ChannelTab, error) {
	return alert.FindAlertChannelByUUID(s.db, uuid)
}

func (s *Service) FindStrategyFromDBByID(strategyID string) (*alert.StrategyTab, error) {
	return alert.FindStrategyByID(s.db, strategyID)
}

func (s *Service) FindAlertCountByCond(cond *request.GetAlertRuleListRequest) (int64, error) {
	var (
		count int64
		err   error
	)

	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.FindAlertCountByCond", sysMetrics.GetStatus(err), time.Since(start))
	count, err = alert.FindAlertCountByCond(s.db, cond)

	return count, err
}

func (s *Service) FindAlertByCond(cond *request.GetAlertRuleListRequest) ([]*alert.RuleAndChannel, error) {
	return alert.FindAlertWithChannelByCond(s.db, cond)
}

func (s *Service) FindAlertMessageListCountByCond(cond *request.GetAlertMessageListRequest) (count int, err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("FindAlertMessageCountByCond", sysMetrics.GetStatus(err), time.Since(start))

	return s.ck.GetAlertMessageCount(&requestMeta.AlertMessageSearch{
		CMDBs:        cond.CMDBs,
		DataBaseName: cond.DBName,
		RuleName:     cond.RuleName,
		Env:          cond.DBEnv,
		IsMute:       cond.IsMute,
		Severity:     cond.Severity,
		Status:       cond.Status,
		TemplateName: cond.TemplateName,
		StartTime:    time.Unix(int64(cond.StartTime), 0),
		EndTime:      time.Unix(int64(cond.EndTime), 0),
	})
}

func (s *Service) FindAlertMessageCountBySpecCond(cond *request.GetAlertMessageListRequest) (count int, err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("FindAlertMessageCountByCond", sysMetrics.GetStatus(err), time.Since(start))

	count, err = s.ck.GetAlertMessageCountBySpecCond(&requestMeta.AlertMessageCountSearch{
		CMDBs:     cond.CMDBs,
		Env:       cond.DBEnv,
		IsMute:    cond.IsMute,
		StartTime: time.Unix(int64(cond.StartTime), 0),
		EndTime:   time.Unix(int64(cond.EndTime), 0),
	})
	return
}

func (s *Service) FindAlertMessageByCond(cond *request.GetAlertMessageListRequest) (resp []*response.AlertMessageWithMuteOperator, err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("FindAlertMessageByCond", sysMetrics.GetStatus(err), time.Since(start))

	return s.ck.GetAlertMessage(&requestMeta.AlertMessageSearch{
		CMDBs:        cond.CMDBs,
		DataBaseName: cond.DBName,
		IsMute:       cond.IsMute,
		RuleName:     cond.RuleName,
		Env:          cond.DBEnv,
		Severity:     cond.Severity,
		Status:       cond.Status,
		TemplateName: cond.TemplateName,
		StartTime:    time.Unix(int64(cond.StartTime), 0),
		EndTime:      time.Unix(int64(cond.EndTime), 0),
		Limit:        cond.PageSize,
		Offset:       (cond.Page - 1) * cond.PageSize,
	})
}

func (s *Service) FindAlertRuleTypeCount(cmdbs []string) (enableNum int, disableNum int, err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("FindAlertRuleTypeCount", sysMetrics.GetStatus(err), time.Since(start))

	enableNum, disableNum, err = alert.FindAlertRuleTypeCount(s.db, cmdbs)
	return
}

func (s *Service) GetAlertMessageByRuleIDAndMessageStatus(ruleID, status string) ([]*response.AlertMessage, error) {
	// 我们需要去查询下 当前message中的获取  rule_id
	return s.ck.GetAlertMessageByMonitorRuleIdAndStatus(ruleID, status)
}
