package alert

import "time"

type MuteConfigManager struct {
	Env            string
	MuteUuid       string
	AlertId        string
	AlertRuleUuid  string
	ChannelUuid    string
	MuteRange      string
	Applicant      string
	MuteTimeRanges []MuteTimeRange
	FilterList     []Filter
}

type MuteTimeRange struct {
	StartTime string `json:"start_time,omitempty"`
	StartUnix int64  `json:"start_unix"`
	EndTime   string `json:"end_time,omitempty"`
	EndUnix   int64  `json:"end_unix"`
	Type      string `json:"type,omitempty"`
}

type Filter struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}

type CreateAlertRuleMuteConfigRequest struct {
	DisplayName   string          `json:"display_name,omitempty"`
	MuteDesc      string          `json:"mute_desc,omitempty"`
	MuteLabelOrg  string          `json:"mute_label_org,omitempty"`
	ProjectName   string          `json:"project_name,omitempty"`
	MuteFilter    MuteFilter      `json:"mute_filter,omitempty"`
	MuteTimeRange []MuteTimeRange `json:"mute_time_range,omitempty"`
	IsDisabled    int             `json:"is_disabled,omitempty"`
}

type GetAlertEventRequest struct {
	AuthProject    string            `json:"auth_project"`
	AuthEventSetId string            `json:"auth_event_set_id"`
	StartTime      int64             `json:"start_time"`
	EndTime        int64             `json:"end_time"`
	ProjectName    string            `json:"project_name"`
	RuleId         string            `json:"rule_id"`
	AlertId        string            `json:"alert_id"`
	Annotations    map[string]string `json:"annotations"`
}

type GetAlertEventResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    EventDatas `json:"data"`
}

type GetAlertEventDetailResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    EventData `json:"data"`
}

type EventDatas struct {
	Total int         `json:"total"`
	Items []EventData `json:"items"`
}

type EventData struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	DisplayStatus string `json:"display_status"`
}

type UpdateAlertRuleMuteConfigRequest struct {
	MuteConfig RuleMuteConfig `json:"mute_config,omitempty"`
	UpdateMask UpdateMask     `json:"update_mask,omitempty"`
}

type MuteFilter struct {
	Version string `json:"version,omitempty"`
	Spec    struct {
		FilterList []Filter `json:"filter,omitempty"`
	} `json:"spec,omitempty"`
}

type CreateAlertRuleMuteConfigResponse struct {
	RuleMuteConfig
}

type UpdateAlertRuleMuteConfigResponse struct {
	RuleMuteConfig
}

type GetSingleAlertRuleMuteConfigResponse struct {
	RuleMuteConfig
}

type RuleMuteConfig struct {
	MuteId        string          `json:"mute_id,omitempty"`
	MuteDesc      string          `json:"mute_desc,omitempty"`
	MuteLabelOrg  string          `json:"mute_label_org,omitempty"`
	ProjectName   string          `json:"project_name,omitempty"`
	MuteFilter    *MuteFilter     `json:"mute_filter,omitempty"`
	MuteTimeRange []MuteTimeRange `json:"mute_time_range,omitempty"`
	DisplayName   string          `json:"display_name,omitempty"`
	DataVersion   string          `json:"data_version,omitempty"`
	Creator       string          `json:"creator,omitempty"`
	Updater       string          `json:"updater,omitempty"`
	CreateTime    time.Time       `json:"create_time,omitempty"`
	UpdateTime    time.Time       `json:"update_time,omitempty"`
	IsDisabled    int             `json:"is_disabled,omitempty"`
}
