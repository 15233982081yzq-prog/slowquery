package response

import "time"

type AlertMute struct {
	MonitorMuteID  string    `gorm:"column:monitor_mute_id"`
	Env            string    `gorm:"column:environment"`
	MuteTitle      string    `gorm:"column:mute_title"`
	RuleUUID       string    `gorm:"column:rule_uuid"`
	MonitorAlertID string    `gorm:"column:monitor_alert_id"`
	Status         string    `gorm:"column:status"`
	MuteFilter     string    `gorm:"column:mute_filter"`
	StartTime      uint64    `gorm:"column:start_time"`
	EndTime        uint64    `gorm:"column:end_time"`
	CreateTime     time.Time `gorm:"column:create_time"`
	Creator        string    `gorm:"column:creator"`
}
