package response

import (
	"time"

	"smart-slowquery/conf"
)

type AlertVo struct {
	AlertUUIDs   []string `json:"alert_uuids,omitempty"`
	ChannelUUIDs []string `json:"channel_uuids,omitempty"`
}

type UpdateAlertVo struct {
	AlertUUID   string `json:"alert_uuid,omitempty"`
	ChannelUUID string `json:"channel_uuid,omitempty"`
}

type BatchUpdateStatusVo struct {
	SuccessRules []string   `json:"success_rules,omitempty"`
	FailRules    []FailInfo `json:"fail_rules,omitempty"`
}

func (b *BatchUpdateStatusVo) FailAlertNames() (names []string) {
	for _, fail := range b.FailRules {
		names = append(names, fail.AlertName)
	}
	return
}

type FailInfo struct {
	UUID      string `json:"uuid"`
	ErrMsg    string `json:"err_msg"`
	AlertName string `json:"-"`
}

type GetTemplatesVo struct {
	Templates []*conf.AlertTemplate `json:"templates,omitempty"`
}

type AlertDetailVo struct {
	RuleUUID    string         `json:"rule_uuid"`
	ChannelUUID string         `json:"channel_uuid"`
	CMDB        string         `json:"cmdb"`
	DBEnv       string         `json:"db_env"`
	DBS         []string       `json:"dbs"`
	Status      string         `json:"status"` // disable enable
	Applicant   string         `json:"applicant"`
	Channel     *ChannelInfo   `json:"channel"`
	Rule        *AlertRuleInfo `json:"rule"`
}

type AlertRuleInfo struct {
	Name              string `json:"name"`
	Trigger           string `json:"trigger"`
	Type              string `json:"type"`
	AlertTemplateUUID string `json:"alert_template_uuid"`
	Expression        string `json:"expression"`
	ExpressionValue   int    `json:"expression_value"`
	ForRange          string `json:"for_range"`
	EvaluateEvery     string `json:"evaluate_every"`
	Severity          string `json:"severity"`
	AlertMsg          string `json:"alert_msg"`
	ResolveMsg        string `json:"resolve_msg"`
}

type ChannelInfo struct {
	Recipient    Recipient    `json:"recipient"`
	Notification Notification `json:"notification"`
}

type Recipient struct {
	Dod      int      `json:"dod"`
	Users    []string `json:"users"`
	Interval string   `json:"interval"`
}

type Notification struct {
	Phone      *Phone      `json:"phone,omitempty"`
	Seatalk    *Seatalk    `json:"seatalk,omitempty"`
	Email      *Email      `json:"email,omitempty"`
	MatterMost *MatterMost `json:"mattermost,omitempty"`
}

type Phone struct {
	CallType int `json:"call_type"`
}

type Seatalk struct {
	WebHook []string `json:"web_hook"`
}

type Email struct{}

type MatterMost struct {
	Channel []string `json:"channel"`
}

type GetAlertRuleListResponse struct {
	TotalNum  int            `json:"total_num"`
	TotalPage int            `json:"total_page"`
	PageSize  int            `json:"page_size"`
	Rules     []*AlertInfoVo `json:"rules"`
}

type GetAlertMessageListResponse struct {
	TotalNum  int                   `json:"total_num"`
	TotalPage int                   `json:"total_page"`
	PageSize  int                   `json:"page_size"`
	Alerts    []*AlertMessageInfoVo `json:"alerts"`
}

type GetAlertMessageAbstractResponse struct {
	CurrentDay   int `json:"current_day"`
	Last10Min    int `json:"last_10_min"`
	MuteNum      int `json:"mute_num"`
	EnableRules  int `json:"enable_rules"`
	DisableRules int `json:"disable_rules"`
}

type AlertInfoVo struct {
	CMDB            string    `json:"cmdb"`
	DBEnv           string    `json:"db_env"`
	DBS             []string  `json:"dbs"`
	Status          string    `json:"status"` // disable enable
	AlertName       string    `json:"alert_name"`
	AlertUUID       string    `json:"alert_uuid"`
	Creator         string    `json:"creator"`
	EvaluateEvery   string    `json:"evaluate_every"`
	Expression      string    `json:"expression"`
	ExpressionValue int       `json:"expression_value"`
	ForRange        string    `json:"for_range"`
	Severity        string    `json:"severity"`
	ChannelType     string    `json:"channel_type"`
	CreateTime      time.Time `json:"create_time"`
	UpdateTime      time.Time `json:"update_time"`
	TemplateName    string    `json:"template_name"`
	Recipient       []string  `json:"recipient"`
}

type AlertMessageInfoVo struct {
	AlertID       string    `json:"alert_id"`
	AlertRuleName string    `json:"alert_rule_name"`
	TemplateName  string    `json:"template_name"`
	CMDB          string    `json:"cmdb"`
	DBName        string    `json:"db_name"`
	Severity      string    `json:"severity"`
	FirsTime      time.Time `json:"firs_time"`
	LastTime      time.Time `json:"last_time"`
	Duration      string    `json:"duration"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	Count         int       `json:"count"`
	AlertStrategy string    `json:"alert_strategy"`
	MuteOperator  string    `json:"mute_operator"`
	MuteTo        int64     `json:"mute_to"`
}

type MuteResponse struct {
	SuccessAlerts []string   `json:"success_alerts"`
	FailAlerts    []FailInfo `json:"failed_alerts"`
}

type AckResponse MuteResponse
