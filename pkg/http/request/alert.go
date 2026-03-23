package request

import (
	"fmt"
)

// AlertCreateRequest 创建报警请求参数
type AlertCreateRequest struct {
	*AlertCommonRequest
}

// AlertUpdateRequest 更新报警请求参数
type AlertUpdateRequest struct {
	RuleUUID    string `json:"rule_uuid" binding:"required"`
	ChannelUUID string `json:"channel_uuid" binding:"required"`
	*AlertCommonRequest
}

type AlertCommonRequest struct {
	CMDB      string         `json:"cmdb" binding:"required"`
	DBEnv     string         `json:"db_env" binding:"required"`
	DBS       []string       `json:"dbs" binding:"required"`
	Status    string         `json:"status"` // 'disable,enable'
	Applicant string         `json:"applicant"`
	Channel   *ChannelInfo   `json:"channel" binding:"required"`
	Rule      *AlertRuleInfo `json:"rule" binding:"required"`
	DBGroup   [][]string     `json:"-"`
}

type Status int

const (
	DISABLE = "disable"
	ENABLE  = "enable"
)

type AlertRuleInfo struct {
	Name              string `json:"name" binding:"required"`
	Trigger           string `json:"trigger" binding:"required,oneof=NOTIFY_WHEN_TRIGGERED_AND_RESOLVED"`
	AlertTemplateUUID string `json:"alert_template_uuid" binding:"required"`
	Expression        string `json:"expression" binding:"required"`
	ExpressionValue   int    `json:"expression_value" binding:"required"`
	ForRange          string `json:"for_range" binding:"required"`
	EvaluateEvery     string `json:"evaluate_every" binding:"required"`
	Severity          string `json:"severity" binding:"required,oneof=warn error critical"`
	AlertMsg          string `json:"alert_msg" binding:"required,max=600"`
	ResolveMsg        string `json:"resolve_msg" binding:"required,max=600"`
}

func (a *AlertRuleInfo) GetAlertStrategy() string {
	return fmt.Sprintf("Evaluate Every %s %s %d For %s", a.EvaluateEvery, a.Expression, a.ExpressionValue, a.ForRange)
}

type ChannelInfo struct {
	Recipient    Recipient    `json:"recipient"`
	Notification Notification `json:"notification"`
}

type Recipient struct {
	Interval string   `json:"interval" binding:"required"`
	Dod      int      `json:"dod"`
	Users    []string `json:"users"`
}

type Notification struct {
	Phone      *Phone      `json:"phone"`
	Seatalk    *Seatalk    `json:"seatalk"`
	Email      *Email      `json:"email"`
	MatterMost *MatterMost `json:"mattermost"`
}

type Phone struct {
	CallType int `json:"call_type"`
}

type Seatalk struct {
	WebHook []string `json:"web_hook"`
}

type Email struct {
}

type MatterMost struct {
	Channel []string `json:"channel"`
}

type ChangeAlertRuleStatusRequest struct {
	Applicant string `json:"applicant"`
	DBEnv     string `json:"db_env" binding:"required"`
	Status    string `json:"status" binding:"required"`
	RuleUuid  string `json:"rule_uuid" binding:"required"`
}

type DeleteRulesRequest struct {
	Applicant string   `json:"applicant"`
	DBEnv     string   `json:"db_env" binding:"required"`
	RuleUuids []string `json:"rule_uuids" binding:"required"`
}

type BatchChangeAlertRuleStatusRequest struct {
	Applicant string   `json:"applicant"`
	Status    string   `json:"status" binding:"required"`
	DBEnv     string   `json:"db_env" binding:"required"`
	RuleUuids []string `json:"rule_uuids" binding:"required"`
}

type GetAlertRuleListRequest struct {
	Page              int      `json:"page" binding:"required"`
	PageSize          int      `json:"page_size" binding:"required"`
	Name              string   `json:"alert_name"`
	Applicant         string   `json:"applicant"`
	CMDBS             []string `json:"cmdbs"`
	Severity          string   `json:"severity"`
	DBEnv             string   `json:"db_env" binding:"required"`
	UserType          string   `json:"user_type"`
	Status            string   `json:"status"`
	AlertTemplateName string   `json:"alert_template_name"`
	Creator           string   `json:"creator"`
	StartTime         int      `json:"start_time"`
	EndTime           int      `json:"end_time"`
}

type GetAlertMessageListRequest struct {
	Page         int      `json:"page" binding:"required"`
	PageSize     int      `json:"page_size" binding:"required"`
	DBEnv        string   `json:"db_env" binding:"required"`
	RuleName     string   `json:"alert_rule_name"`
	Applicant    string   `json:"applicant"`
	CMDBs        []string `json:"cmdbs"`
	DBName       string   `json:"db_name"`
	Status       string   `json:"status"`
	Severity     string   `json:"severity"`
	TemplateName string   `json:"template_name"`
	IsMute       bool     `json:"is_mute"`
	StartTime    int      `json:"start_time"`
	EndTime      int      `json:"end_time"`
}

type GetAlertMessageAbstractRequest struct {
	DBEnv     string   `json:"db_env" binding:"required"`
	Applicant string   `json:"applicant"`
	CMDBs     []string `json:"cmdbs"`
}

type CreateMuteRequest struct {
	AlertIds  []string `json:"alert_ids" binding:"required,max=20"`
	Applicant string   `json:"applicant"`
	MuteTime  string   `json:"mute_time" binding:"required"`
	DBEnv     string   `json:"db_env" binding:"required"`
}

type CancelMuteRequest struct {
	AlertIds  []string `json:"alert_ids" binding:"required,max=20"`
	Applicant string   `json:"applicant"`
	DBEnv     string   `json:"db_env" binding:"required"`
}

type CreateAckRequest struct {
	AlertIds  []string `json:"alert_ids" binding:"required,max=20"`
	Applicant string   `json:"applicant"`
	DBEnv     string   `json:"db_env" binding:"required"`
}
