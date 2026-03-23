package request

import (
	"smart-slowquery/pkg/oplog"
	"smart-slowquery/pkg/store/request"
)

type AlertCallBackRequest struct {
	AlertRuleId       string       `json:"alert_rule_id"`
	AlertRuleUuid     string       `json:"alert_rule_uuid"`
	ChannelUuid       string       `json:"channel_uuid"`
	AlertStrategy     string       `json:"alert_strategy"`
	ChannelName       string       `json:"channel_name"`
	AlertName         string       `json:"alert_name"`
	Message           string       `json:"message"`
	AlertTemplateUUID string       `json:"alert_template_uuid"`
	AlertList         []*AlertInfo `json:"alert_list"`
	From              string       `json:"from"`
}

func (req *AlertCallBackRequest) HasMessageClosed() (bl bool) {
	for _, alertInfo := range req.AlertList {
		if alertInfo.Status == request.MonitorClosedStatus {
			return true
		}
	}
	return false
}

func (req *AlertCallBackRequest) FromMonitor() bool {
	return req.From == oplog.MonitorPlatform
}

func (req *AlertCallBackRequest) FromSlowQuery() bool {
	return req.From == oplog.SlowQueryAlertPlatform
}

type AlertInfo struct {
	AlertId       int64       `json:"alert_id,omitempty"`
	StartTime     uint64      `json:"start_time,omitempty"`
	AlertRuleId   string      `json:"alert_rule_id,omitempty"`
	AlertRuleName string      `json:"alert_rule_name,omitempty"`
	Status        string      `json:"status,omitempty"`
	Severity      string      `json:"severity,omitempty"`
	Message       string      `json:"message,omitempty"`
	AckInfo       string      `json:"ack_info,omitempty"`
	DeployEnv     string      `json:"deploy_env,omitempty"`
	Labels        *AlertLabel `json:"labels,omitempty"`
}

type AlertLabel struct {
	AlertGroup   string `json:"alertgroup,omitempty"`
	AlertName    string `json:"alertname,omitempty"`
	DatabaseName string `json:"database_name,omitempty"`
	CMDB         string `json:"cmdb,omitempty"`
	DbEnv        string `json:"db_env,omitempty"`
	MRuleID      string `json:"m_rule_id"`
}
