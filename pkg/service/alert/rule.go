package alert

import (
	requestMeta "smart-slowquery/pkg/http/request"

	"time"
)

// CreateAlertRuleRequest 调用monitor创建rule参数
type CreateAlertRuleRequest struct {
	RuleLabelOrg                   string          `json:"rule_label_org,omitempty"`
	ProjectName                    string          `json:"project_name,omitempty"`
	MetricStoreNames               [][]string      `json:"metric_store_names,omitempty"`
	RuleType                       string          `json:"rule_type,omitempty"`
	DisplayName                    string          `json:"display_name,omitempty"`
	RuleDesc                       string          `json:"rule_desc,omitempty"`
	AlertRuleContent               RuleContent     `json:"alert_rule_content,omitempty"`
	AlertRuleLabels                RuleLabels      `json:"alert_rule_labels,omitempty"`
	GrafanaUrl                     string          `json:"grafana_url,omitempty"`
	IsDisabled                     int             `json:"is_disabled,omitempty"`
	IsUseProjectDefaultMetricStore bool            `json:"is_use_project_default_metric_store,omitempty"`
	NotificationType               string          `json:"notification_type,omitempty"`
	ChannelConfig                  []ChannelConfig `json:"channel_config,omitempty"`
}

type RuleContent struct {
	Severity string  `json:"severity,omitempty"`
	Content  Content `json:"content,omitempty"`
}

type RuleLabels struct {
	Source        string `json:"source"`
	CMDB          string `json:"cmdb"`
	AlertRuleUuid string `json:"alert_rule_uuid"`
	DBEnv         string `json:"db_env"`
}

type Content struct {
	For                string          `json:"for,omitempty"`
	EvaluationInterval string          `json:"evaluation_interval,omitempty"`
	Expr               string          `json:"expr,omitempty"`
	MessageTemplate    MessageTemplate `json:"message_template,omitempty"`
}

type MessageTemplate struct {
	FiringMessage   string `json:"firing_message,omitempty"`
	ResolvedMessage string `json:"resolved_message,omitempty"`
}

type ChannelConfig struct {
	Names  []string             `json:"names,omitempty"`
	Config *ChannelConfigDetail `json:"config,omitempty"`
}

type ChannelConfigDetail struct {
	Mattermost    *MattermostConfig    `json:"mattermost,omitempty"`
	Tele          *TeleConfig          `json:"tele,omitempty"`
	Seatalk       *SeatalkConfig       `json:"seatalk,omitempty"`
	Email         *EmailConfig         `json:"email,omitempty"`
	Sms           *SmsConfig           `json:"sms,omitempty"`
	GoogleChatbot *GoogleChatbotConfig `json:"google_chatbot,omitempty"`
	Webhook       *WebhookConfig       `json:"webhook,omitempty"`
}

type TeleConfig struct {
	ChannelConfigCommon
	DialingOrder int `json:"dialing_order"`
}

type SeatalkConfig struct {
	ChannelConfigCommon
}

type MattermostConfig struct {
	ChannelConfigCommon
}

type EmailConfig struct {
	ChannelConfigCommon
}

type SmsConfig struct {
	ChannelConfigCommon
}

type WebhookConfig struct {
	ChannelConfigCommon
}

type GoogleChatbotConfig struct {
	ChannelConfigCommon
}

type ChannelConfigCommon struct {
	RepeatInterval    string `json:"repeat_interval"`
	MessageTemplateId string `json:"message_template_id"`
	Enabled           bool   `json:"enabled"`
}

type CreateAlertRuleResponse struct {
	Rule
}

type GetSingleAlertRuleResponse struct {
	Rule
}

type Rule struct {
	RuleId                         string          `json:"rule_id,omitempty"`
	RuleLabelOrg                   string          `json:"rule_label_org,omitempty"`
	ProjectName                    string          `json:"project_name,omitempty"`
	MetricStoreNames               [][]string      `json:"metric_store_names,omitempty"`
	AlertRuleContent               *RuleContent    `json:"alert_rule_content,omitempty"`
	AlertRuleLabels                *RuleLabels     `json:"alert_rule_labels,omitempty"`
	IsDisabled                     *int            `json:"is_disabled,omitempty"`
	IsUseProjectDefaultMetricStore bool            `json:"is_use_project_default_metric_store,omitempty"`
	RuleType                       string          `json:"rule_type,omitempty"`
	DisplayName                    string          `json:"display_name,omitempty"`
	RuleDesc                       string          `json:"rule_desc,omitempty"`
	GrafanaUrl                     string          `json:"grafana_url,omitempty"`
	DataVersion                    string          `json:"data_version,omitempty"`
	Creator                        string          `json:"creator,omitempty"`
	Updater                        string          `json:"updater,omitempty"`
	CreateTime                     time.Time       `json:"create_time,omitempty"`
	UpdateTime                     time.Time       `json:"update_time,omitempty"`
	DirectBindId                   string          `json:"direct_bind_id,omitempty"`
	NotificationType               string          `json:"notification_type,omitempty"`
	ChannelConfig                  []ChannelConfig `json:"channel_config,omitempty"`
	IsDelete                       bool            `json:"-"`
}

func (r *Rule) Disable() bool {
	return *r.IsDisabled == 1
}

func (r *Rule) Status() string {
	if r.Disable() {
		return requestMeta.DISABLE
	}
	return requestMeta.ENABLE
}

type UpdateAlertRuleRequest struct {
	AlertRule  Rule       `json:"alert_rule,omitempty"`
	UpdateMask UpdateMask `json:"update_mask,omitempty"`
}

type UpdateAlertRuleResponse struct {
	Rule
	ChannelConfig []ChannelConfig `json:"channel_config,omitempty"`
}
