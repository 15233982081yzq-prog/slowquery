package alert

import "time"

// CreateNotifyChannelRequest 调用monitor创建channel参数
type CreateNotifyChannelRequest struct {
	ChannelName     string  `json:"channel_name,omitempty"`
	DisplayName     string  `json:"display_name,omitempty"`
	Channel         Channel `json:"channel_config,omitempty"`
	IsDisabled      int     `json:"is_disabled,omitempty"`
	ChannelDesc     string  `json:"channel_desc,omitempty"`
	ChannelLabelOrg string  `json:"channel_label_org,omitempty"`
	ProjectName     string  `json:"project_name,omitempty"`
}

// CreateNotifyChannelResponse 调用monitor返回结构
type CreateNotifyChannelResponse struct {
	NotifyChannel
}

type GetSingleNotifyChannelResponse struct {
	NotifyChannel
}

// NotifyChannel 调用monitor返回详细结构
type NotifyChannel struct {
	ChannelName     string    `json:"channel_name,omitempty"`
	ChannelLabelOrg string    `json:"channel_label_org,omitempty"`
	DisplayName     string    `json:"display_name,omitempty"`
	ProjectName     string    `json:"project_name,omitempty"`
	ChannelDesc     string    `json:"channel_desc,omitempty"`
	Channel         Channel   `json:"channel_config,omitempty"`
	IsDisabled      *int      `json:"is_disabled,omitempty"`
	DataVersion     string    `json:"data_version,omitempty"`
	Creator         string    `json:"creator,omitempty"`
	Updater         string    `json:"updater,omitempty"`
	CreateTime      time.Time `json:"create_time,omitempty"`
	UpdateTime      time.Time `json:"update_time,omitempty"`
	IsDelete        bool      `json:"-"`
}

func (nc *NotifyChannel) Disable() bool {
	return *nc.IsDisabled == 1
}

type Channel struct {
	Version string      `json:"version,omitempty"`
	Spec    ChannelSpec `json:"spec,omitempty"`
}

type ChannelSpec struct {
	ChannelInfo ChannelInfo `json:"channel,omitempty"`
}

type ChannelInfo struct {
	Mattermost    *MattermostInfo    `json:"mattermost,omitempty"`
	Tele          *TeleInfo          `json:"tele,omitempty"`
	Seatalk       *SeatalkInfo       `json:"seatalk,omitempty"`
	Email         *EmailInfo         `json:"email,omitempty"`
	Sms           *SmsInfo           `json:"sms,omitempty"`
	GoogleChatbot *GoogleChatbotInfo `json:"google_chatbot,omitempty"`
	Webhook       *WebhookInfo       `json:"webhook,omitempty"`
}

type MattermostInfo struct {
	SendTo []MattermostRecipient `json:"send_to,omitempty"`
}

type MattermostRecipient struct {
	MmChannel  string   `json:"mm_channel,omitempty"`
	MmMentions []string `json:"mm_mentions,omitempty"`
	MmDesc     string   `json:"mm_desc,omitempty"`
}

type TeleInfo struct {
	SendTo []string `json:"send_to,omitempty"`
}

type SeatalkInfo struct {
	SendTo []SeatalkRecipient `json:"send_to,omitempty"`
}

type SeatalkRecipient struct {
	BotWebhookUrl string   `json:"bot_webhook_url,omitempty"`
	BotMentions   []string `json:"bot_mentions,omitempty"`
	BotDesc       string   `json:"bot_desc,omitempty"`
	AtAll         bool     `json:"at_all,omitempty"`
}

type EmailInfo struct {
	SendTo []string `json:"send_to,omitempty"`
}

type SmsInfo struct {
	SendTo []string `json:"send_to,omitempty"`
}

type GoogleChatbotInfo struct {
	SendTo []GoogleChatbotRecipient `json:"send_to,omitempty"`
}

type GoogleChatbotRecipient struct {
	SpaceId         string   `json:"space_id,omitempty"`
	ChatbotMentions []string `json:"chatbot_mentions,omitempty"`
	ChatbotDesc     string   `json:"chatbot_desc,omitempty"`
}

type WebhookInfo struct {
	SendTo []WebhookRecipient `json:"send_to,omitempty"`
}

func (o *WebhookInfo) AddWebhookRecipient(v WebhookRecipient) {
	o.SendTo = append(o.SendTo, v)
}

type WebhookRecipient struct {
	WebhookUrl    string          `json:"webhook_url,omitempty"`
	WebhookHeader []WebhookHeader `json:"webhook_header,omitempty"`
	WebhookBody   string          `json:"webhook_body,omitempty"`
	WebhookDesc   string          `json:"webhook_desc,omitempty"`
}

type WebhookHeader struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type UpdateNotifyChannelRequest struct {
	NotifyChannel NotifyChannel `json:"notify_channel,omitempty"`
	UpdateMask    UpdateMask    `json:"update_mask,omitempty"`
}

type DeleteNotifyChannelRequest struct {
	ProjectName string `json:"project_name"`
}

type UpdateMask struct {
	Paths []string `json:"paths,omitempty"`
}

type UpdateNotifyChannelResponse struct {
	NotifyChannel
}
