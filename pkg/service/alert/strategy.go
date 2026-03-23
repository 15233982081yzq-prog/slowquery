package alert

import "time"

type CreateNotificationStrategyRequest struct {
	BindLabelOrg      string          `json:"bind_label_org,omitempty"`
	ProjectName       string          `json:"project_name,omitempty"`
	DisplayName       string          `json:"display_name,omitempty"`
	BindFilter        [][]Bind        `json:"bind_filter,omitempty"`
	EnableResolvedMsg *bool           `json:"enable_resolved_msg,omitempty"`
	IsDisabled        int             `json:"is_disabled"`
	ChannelConfig     []ChannelConfig `json:"channel_config,omitempty"`
	StrategyType      string          `json:"strategy_type,omitempty"` // must:GENERAL
}

type Bind struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}

type CreateNotificationStrategyResponse struct {
	StrategyResponse
}

type GetSingleNotificationStrategyResponse struct {
	BindId            string          `json:"bind_id"`
	BindLabelOrg      string          `json:"bind_label_org"`
	ProjectName       string          `json:"project_name"`
	BindFilter        [][]Bind        `json:"bind_filter"`
	IsDisabled        *int            `json:"is_disabled"`
	DisplayName       string          `json:"display_name"`
	ChannelConfig     []ChannelConfig `json:"channel_config"`
	EnableResolvedMsg bool            `json:"enable_resolved_msg"`
	GroupLabels       []string        `json:"group_labels"`
	DataVersion       string          `json:"data_version"`
	Creator           string          `json:"creator"`
	Updater           string          `json:"updater"`
	CreateTime        time.Time       `json:"create_time"`
	UpdateTime        time.Time       `json:"update_time"`
	StrategyType      string          `json:"strategy_type"`
	IsDelete          bool            `json:"-"`
}

func (strategy *GetSingleNotificationStrategyResponse) Disable() bool {
	return *strategy.IsDisabled == 1
}

type DeleteNotificationStrategyRequest struct {
	ProjectName string `json:"project_name"`
}

type UpdateNotificationStrategyRequest struct {
	AlertNotificationBind NotificationBind `json:"alert_notification_bind"`
	UpdateMask            UpdateMask       `json:"update_mask"`
}

type NotificationBind struct {
	DisplayName       string          `json:"display_name,omitempty"`
	BindFilter        [][]Bind        `json:"bind_filter,omitempty"`
	EnableResolvedMsg bool            `json:"enable_resolved_msg,omitempty"`
	IsDisabled        *int            `json:"is_disabled,omitempty"`
	DataVersion       string          `json:"data_version,omitempty"`
	ChannelConfig     []ChannelConfig `json:"channel_config,omitempty"`
}

type StrategyResponse struct {
	BindId            string          `json:"bind_id"`
	BindLabelOrg      string          `json:"bind_label_org"`
	ProjectName       string          `json:"project_name"`
	BindFilter        [][]Bind        `json:"bind_filter"`
	IsDisabled        int             `json:"is_disabled"`
	DisplayName       string          `json:"display_name"`
	ChannelConfig     []ChannelConfig `json:"channel_config"`
	EnableResolvedMsg bool            `json:"enable_resolved_msg"`
	GroupLabels       []string        `json:"group_labels"`
	DataVersion       string          `json:"data_version"`
	Creator           string          `json:"creator"`
	Updater           string          `json:"updater"`
	CreateTime        time.Time       `json:"create_time"`
	UpdateTime        time.Time       `json:"update_time"`
	StrategyType      string          `json:"strategy_type"`
}

type UpdateNotificationStrategyResponse struct {
	StrategyResponse
}
