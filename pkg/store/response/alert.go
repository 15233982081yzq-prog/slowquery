package response

import "time"

type AlertMessage struct {
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

type AlertMessageWithMuteOperator struct {
	AlertMessage
	MuteOperator string `gorm:"column:mute_operator"`
	MuteTo       int64  `gorm:"column:mute_to"`
}
