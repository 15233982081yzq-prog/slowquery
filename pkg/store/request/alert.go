package request

import (
	"fmt"
	"time"
)

type AlertOperatorLog struct {
	BaseRequest
	Operator   string    `gorm:"column:operator"`
	ActionID   string    `gorm:"column:action_id"`
	ActionType string    `gorm:"column:action_type"`
	ActionName string    `gorm:"column:action_name"`
	Env        string    `gorm:"column:environment"`
	CreateTime time.Time `gorm:"column:create_time"`
	OldSetting string    `gorm:"column:old_setting"`
	NewSetting string    `gorm:"column:new_setting"`
}

func BuildAlertOperatorLog(operator, actionID, actionType, actionName, env, oldSetting, newSetting string) *AlertOperatorLog {
	return &AlertOperatorLog{
		Operator:   operator,
		ActionID:   actionID,
		ActionType: actionType,
		ActionName: actionName,
		Env:        env,
		OldSetting: oldSetting,
		NewSetting: newSetting,
		CreateTime: time.Now(),
	}
}

func (req *AlertOperatorLog) Valid() error {
	if len(req.Operator) == 0 {
		return fmt.Errorf("AlertOperatorLog param failed, operator is empty")
	}
	if len(req.ActionID) == 0 {
		return fmt.Errorf("AlertOperatorLog param failed, actionID is empty")
	}
	if len(req.ActionType) == 0 {
		return fmt.Errorf("AlertOperatorLog param failed, actionType is empty")
	}
	if len(req.ActionName) == 0 {
		return fmt.Errorf("AlertOperatorLog param failed, actionName is empty")
	}
	if len(req.Env) == 0 {
		return fmt.Errorf("AlertOperatorLog param failed, env is empty")
	}
	return nil
}

func (req *AlertOperatorLog) TableName() string {
	return "alert_operation_log_all_rand"
}

type AlertMessage struct {
	BaseRequest
	MonitorAlertID string    `gorm:"column:monitor_alert_id"`
	MonitorRuleID  string    `gorm:"column:monitor_rule_id"`
	AlertRuleUUID  string    `gorm:"column:alert_rule_uuid"`
	AlertStrategy  string    `gorm:"column:alert_strategy"`
	AlertRuleName  string    `gorm:"column:alert_rule_name"`
	ChannelUUID    string    `gorm:"column:channel_uuid"`
	CMDB           string    `gorm:"column:cmdb"`
	DataBaseName   string    `gorm:"column:database_name"`
	Env            string    `gorm:"column:environment"`
	Status         string    `gorm:"column:alert_status"`
	Severity       string    `gorm:"column:severity"`
	Message        string    `gorm:"column:message"`
	ACKBy          string    `gorm:"column:ack_by"` // 记录最后一次ack的人员邮箱
	LabelInfo      string    `gorm:"column:label_info"`
	TemplateName   string    `gorm:"column:template_name"`
	AlertCount     uint64    `gorm:"column:alert_count"`
	StartTime      uint64    `gorm:"column:start_time"`
	ResolveTime    uint64    `gorm:"column:resolve_time"`
	LastAlertTime  time.Time `gorm:"column:last_alert_time"`
	CreateTime     time.Time `gorm:"column:create_time"`
	UpdateTime     time.Time `gorm:"column:update_time"`
}

func BuildAlertMessage(monitorAlertID, monitorRuleID, alertRuleUUID, channelUUID, cmdb, dataBaseName, env, status, severity, message, ackBy, labelInfo, templateName string, alertCount, startTime, resolveTime uint64, lastAlertTime time.Time) *AlertMessage {
	return &AlertMessage{
		MonitorAlertID: monitorAlertID,
		MonitorRuleID:  monitorRuleID,
		AlertRuleUUID:  alertRuleUUID,
		ChannelUUID:    channelUUID,
		CMDB:           cmdb,
		DataBaseName:   dataBaseName,
		Env:            env,
		Status:         status,
		Severity:       severity,
		Message:        message,
		ACKBy:          ackBy,
		LabelInfo:      labelInfo,
		TemplateName:   templateName,
		AlertCount:     alertCount,
		StartTime:      startTime,
		ResolveTime:    resolveTime,
		LastAlertTime:  lastAlertTime,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
}

const (
	AlertResolved = "resolved"
	AlertClosed   = "closed"
	AlertPending  = "pending"

	MonitorActiveStatus   = "firing"
	MonitorClosedStatus   = "closed"
	MonitorResolvedStatus = "resolved"
)

func (req *AlertMessage) Valid() error {
	if len(req.MonitorAlertID) == 0 {
		return fmt.Errorf("AlertMessage param failed, MonitorAlertID is empty")
	}
	if len(req.MonitorRuleID) == 0 {
		return fmt.Errorf("AlertMessage param failed, MonitorRuleID is empty")
	}
	if len(req.AlertRuleUUID) == 0 {
		return fmt.Errorf("AlertMessage param failed, AlertRuleUUID is empty")
	}
	if len(req.ChannelUUID) == 0 {
		return fmt.Errorf("AlertMessage param failed, ChannelUUID is empty")
	}
	if len(req.CMDB) == 0 {
		return fmt.Errorf("AlertMessage param failed, CMDB is empty")
	}
	if len(req.DataBaseName) == 0 {
		return fmt.Errorf("AlertMessage param failed, DataBaseName is empty")
	}
	if len(req.Env) == 0 {
		return fmt.Errorf("AlertMessage param failed, Env is empty")
	}
	if len(req.Status) == 0 {
		return fmt.Errorf("AlertMessage param failed, Status is empty")
	}
	if len(req.Severity) == 0 {
		return fmt.Errorf("AlertMessage param failed, Severity is empty")
	}
	if len(req.Message) == 0 {
		return fmt.Errorf("AlertMessage param failed, Message is empty")
	}
	if len(req.LabelInfo) == 0 {
		return fmt.Errorf("AlertMessage param failed, LabelInfo is empty")
	}
	if len(req.TemplateName) == 0 {
		return fmt.Errorf("AlertMessage param failed, AlertStatus is empty")
	}
	if req.StartTime <= 0 {
		return fmt.Errorf("AlertMessage param failed, startTime <= 0")
	}
	if req.ResolveTime > 0 && req.ResolveTime <= req.StartTime {
		return fmt.Errorf("AlertMessage param failed, startTime >= endTime")
	}
	return nil
}

func (req *AlertMessage) TableName() string {
	return "alert_message_log_all_rand"
}

func (req *AlertMessage) LocalTableName() string {
	return "alert_message_log_local"
}

type QueryType string

type AlertMessageSearch struct {
	BaseRequest
	CMDBs        []string
	DataBaseName string
	Env          string
	Severity     string
	Status       string
	RuleName     string
	TemplateName string
	StartTime    time.Time
	EndTime      time.Time
	IsMute       bool
	Limit        int
	Offset       int
}

type AlertOnMuteMessage struct{}
type AlertNoMutedMessage struct{}

type AlertMessageCountSearch struct {
	BaseRequest
	CMDBs     []string
	Env       string
	StartTime time.Time
	EndTime   time.Time
	IsMute    bool
}

type UpdateStatus struct {
	BaseRequest
	AlertId string
	Status  string
}

func BuildAlertMessageSearch(cmdbs []string, dbname, env, templateName, severity, status string, limit, offset int, start, end time.Time) *AlertMessageSearch {
	return &AlertMessageSearch{
		CMDBs:        cmdbs,
		DataBaseName: dbname,
		Env:          env,
		Severity:     severity,
		Status:       status,
		StartTime:    start,
		EndTime:      end,
		TemplateName: templateName,
		Limit:        limit,
		Offset:       offset,
	}
}

func (req *AlertMessageSearch) Valid() error {
	if len(req.Env) == 0 {
		return fmt.Errorf("AlertMessageSearch param failed, Env is empty")
	}
	if req.StartTime.Unix() > req.EndTime.Unix() {
		return fmt.Errorf("AlertMessageSearch param failed, startTime > endTime")
	}
	if req.Limit < 0 || req.Limit > 50 {
		return fmt.Errorf("AlertMessageSearch param failed, limit:%d max 50", req.Limit)
	}
	return nil
}

func (req *AlertMessageSearch) TableName() string {
	return "alert_message_log_all_rand"
}

func (req *AlertMessageCountSearch) TableName() string {
	return "alert_message_log_all_rand"
}

func (req *UpdateStatus) TableName() string {
	return "alert_message_log_all_rand"
}

func (req *AlertOnMuteMessage) TableName() string {
	return "alert_message_log_all_rand"
}

func (req *AlertNoMutedMessage) TableName() string {
	return "alert_message_log_all_rand"
}

type AlertMessageSummarySearch struct {
	BaseRequest
	CMDBs     []string
	Env       string
	StartTime time.Time
	EndTime   time.Time
}

func BuildAlertMessageSummarySearch(env string, cmdbs []string, start, end time.Time) *AlertMessageSummarySearch {
	return &AlertMessageSummarySearch{
		CMDBs:     cmdbs,
		Env:       env,
		StartTime: start,
		EndTime:   end,
	}
}

func (req *AlertMessageSummarySearch) Valid() error {
	if len(req.Env) == 0 {
		return fmt.Errorf("AlertMessageSearch param failed, Env is empty")
	}
	if req.StartTime.Unix() > req.EndTime.Unix() {
		return fmt.Errorf("AlertMessageSearch param failed, startTime > endTime")
	}
	return nil
}

func (req *AlertMessageSummarySearch) TableName() string {
	return "alert_operation_log_all_rand"
}

func (req *AlertMessageSummarySearch) LocalTableName() string {
	return "alert_operation_log_local"
}

type AlertMute struct {
	BaseRequest
	MonitorMuteID  string    `gorm:"column:monitor_mute_id"`
	Env            string    `gorm:"column:environment"`
	MuteTitle      string    `gorm:"column:mute_title"`
	RuleUUID       string    `gorm:"column:rule_uuid"`
	MonitorAlertID string    `gorm:"column:monitor_alert_id"`
	Status         string    `gorm:"column:status"`
	MuteFilter     string    `gorm:"column:mute_filter"`
	StartTime      uint64    `gorm:"column:start_time"`
	EndTime        uint64    `gorm:"column:end_time"`
	Creator        string    `gorm:"column:creator"`
	CreateTime     time.Time `gorm:"column:create_time"`
	UpdateTime     time.Time `gorm:"column:update_time"`
}

func BuildAlertMute(monitorMuteID, env, muteTitle, ruleUUID, monitorAlertID, status, muteFilter, creator string, startTime, endTime uint64) *AlertMute {
	return &AlertMute{
		MonitorMuteID:  monitorMuteID,
		Env:            env,
		MuteTitle:      muteTitle,
		RuleUUID:       ruleUUID,
		MonitorAlertID: monitorAlertID,
		Status:         status,
		MuteFilter:     muteFilter,
		Creator:        creator,
		StartTime:      startTime,
		EndTime:        endTime,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
}

func (req *AlertMute) Valid() error {
	if len(req.MonitorMuteID) == 0 {
		return fmt.Errorf("AlertMute param failed, monitor_mute_id is empty")
	}
	if len(req.Env) == 0 {
		return fmt.Errorf("AlertMute param failed, environment is empty")
	}
	if len(req.MuteTitle) == 0 {
		return fmt.Errorf("AlertMute param failed, mute_title is empty")
	}
	if len(req.RuleUUID) == 0 {
		return fmt.Errorf("AlertMute param failed, rule_uuid is empty")
	}
	if len(req.MonitorAlertID) == 0 {
		return fmt.Errorf("AlertMute param failed, monitor_alert_id is empty")
	}
	if len(req.Status) == 0 {
		return fmt.Errorf("AlertMute param failed, status is empty")
	}
	if len(req.MuteFilter) == 0 {
		return fmt.Errorf("AlertMute param failed, MuteFilter is empty")
	}
	if req.StartTime <= 0 {
		return fmt.Errorf("AlertMute param failed, startTime <= 0")
	}
	if req.EndTime > 0 && req.EndTime < req.StartTime {
		return fmt.Errorf("AlertMessageSearch param failed, startTime > endTime")
	}
	return nil
}

func (req *AlertMute) TableName() string {
	return "alert_message_mute_all_rand"
}

func (req *AlertMute) LocalTableName() string {
	return "alert_message_mute_local"
}
