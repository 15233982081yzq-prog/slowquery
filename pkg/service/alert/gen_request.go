package alert

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/log"
)

var statusMap = map[string]int{
	request.ENABLE:  0,
	request.DISABLE: 1,
}

const (
	link        = "https://space.shopee.io/console/rds/slow_query/alert/live?name=%s"
	MessageLink = ",slow query alert db:{{$labels.database_name}},number of slow queries in the last hour:{{$value}} > %d,sliding window calculation, statistical period 10 minutes.The alert detail link: " + link
)

var (
	EnableResolvedMsg = true
)

const (
	hookBody = `
	{
        "alert_rule_id":  "%s",
		"alert_rule_uuid":"%s",
		"channel_uuid":   "%s",
		"channel_name":   "%s",
		"alert_template_uuid":  "%s",
		"alert_name":  "%s",
		"alert_strategy": "%s",
		"from": "monitor_platform",
		"message":       "{{$message}}",
		"alert_list":[
			{{range .Events}}
			{
			"alert_id":       {{.ID}},
			"alert_rule_id":   "{{.RuleID}}",
			"alert_rule_name": "{{.Name}}",
			"start_time":     {{.StartTime}},
			"status":        "{{.Status}}",
			"severity":      "{{.Severity}}",
			"message":       "{{.Message}}",
			"labels": {
				"alertgroup": "{{.Labels.alertgroup}}",
				"alertname": "{{.Labels.alertname}}",
				"database_name": "{{.Labels.database_name}}",
				"cmdb": "{{.Annotations.cmdb}}",
				"db_env": "{{.Annotations.db_env}}",
				"m_rule_id": "{{.RuleID}}"
				}
			},
			{{end}}
			{}
		]
	}`

	SstUpLater = "SET_UP_LATER"
)

// GenMgrRuleRequest 构建RuleManager请求结构
func (s *Service) GenMgrRuleRequest(req *request.AlertCommonRequest) ([]*RuleManager, error) {
	var (
		err           error
		promQL        string
		dods          []string
		ok            bool
		alertTemplate *conf.AlertTemplate
		ruleManagers  = make([]*RuleManager, 0)
	)

	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("alertSrv.GenMgrRuleRequest", sysMetrics.GetStatus(err), time.Since(start))
	}()

	// duration time check
	if _, err = time.ParseDuration(req.Rule.ForRange); err != nil {
		log.Errorf("service GenMgrRuleRequest ParseDuration error:%s", err.Error())
		return nil, err
	}
	if _, err = time.ParseDuration(req.Rule.EvaluateEvery); err != nil {
		log.Errorf("service GenMgrRuleRequest ParseDuration error:%s", err.Error())
		return nil, err
	}

	if alertTemplate, ok = s.conf.AlertTemplates[req.Rule.AlertTemplateUUID]; !ok {
		log.Errorf("service AlertTemplates %s not exist", req.Rule.AlertTemplateUUID)
		return nil, err
	}

	for _, dbList := range req.DBGroup {
		singleReq := &request.AlertCommonRequest{
			CMDB:      req.CMDB,
			DBEnv:     req.DBEnv,
			DBS:       dbList,
			Status:    req.Status,
			Applicant: req.Applicant,
			Channel:   req.Channel,
			Rule:      req.Rule,
		}
		if promQL, err = generatePromQL(alertTemplate, &TemplateParam{
			DBS:             dbList,
			Expression:      req.Rule.Expression,
			ExpressionValue: req.Rule.ExpressionValue,
		}); err != nil {
			log.Errorf("service GenMgrRuleRequest generatePromQL error:%s", err.Error())
			return nil, err
		}

		// fetch dod user list. dod 与 users 在传来是时互斥的
		if req.Channel.Recipient.Dod != 0 {
			if dods, err = s.dodClient.ListDodsByTeamID(req.Channel.Recipient.Dod); err != nil {
				log.Errorf("service GenMgrRuleRequest ListDodsByTeamID error:%s", err.Error())
				return nil, err
			}
			req.Channel.Recipient.Users = dods
		}
		ruleManagers = append(ruleManagers, &RuleManager{
			req:    singleReq,
			PromQL: promQL,
		})
	}

	return ruleManagers, nil
}

// 构建报警规则请求结构
func (s *Service) genCreateAlertRuleRequest(ruleManager *RuleManager, alertRuleUuid, channelUuid string) *CreateAlertRuleRequest {

	req := &CreateAlertRuleRequest{
		RuleLabelOrg:     LabelOrg,
		ProjectName:      ProjectName,
		MetricStoreNames: [][]string{{s.conf.MonitorClientConfig.MetricStoreNames}},
		RuleType:         RuleType,
		DisplayName:      ruleManager.req.Rule.Name,
		RuleDesc:         fmt.Sprintf(RuleDesc, alertRuleUuid, channelUuid),
		AlertRuleContent: RuleContent{
			Severity: ruleManager.req.Rule.Severity,
			Content: Content{
				For:                ruleManager.req.Rule.ForRange,
				EvaluationInterval: ruleManager.req.Rule.EvaluateEvery,
				Expr:               ruleManager.PromQL,
				MessageTemplate: MessageTemplate{
					FiringMessage:   ruleManager.req.Rule.AlertMsg + fmt.Sprintf(MessageLink, ruleManager.req.Rule.ExpressionValue, url.QueryEscape(ruleManager.req.Rule.Name)),
					ResolvedMessage: ruleManager.req.Rule.ResolveMsg + fmt.Sprintf(MessageLink, ruleManager.req.Rule.ExpressionValue, url.QueryEscape(ruleManager.req.Rule.Name)),
				},
			},
		},
		AlertRuleLabels: RuleLabels{
			Source:        Source,
			CMDB:          ruleManager.req.CMDB,
			AlertRuleUuid: alertRuleUuid,
			DBEnv:         ruleManager.req.DBEnv,
		},
		GrafanaUrl:                     "", // 可以补充
		IsDisabled:                     statusMap[ruleManager.req.Status],
		IsUseProjectDefaultMetricStore: false,
		NotificationType:               SstUpLater, // ruleManager.req.Rule.Trigger
	}
	return req
}

// 构建channel请求结构
func (s *Service) genCreateChannelRequest(ruleManager *RuleManager, monitorRuleId, alertRuleUuid, channelUuid string) *CreateNotifyChannelRequest {
	channelName := getChannelName(channelUuid)

	req := &CreateNotifyChannelRequest{
		ChannelName: channelName,
		DisplayName: channelName,
		Channel: Channel{
			Version: ChannelVersion,
			Spec:    ChannelSpec{ChannelInfo: ChannelInfo{}},
		},
		IsDisabled:      statusMap[ruleManager.req.Status],
		ChannelDesc:     fmt.Sprintf(ChannelDesc, alertRuleUuid, channelUuid),
		ChannelLabelOrg: LabelOrg,
		ProjectName:     ProjectName,
	}

	// deal channel start
	if ruleManager.req.Channel.Notification.MatterMost != nil {
		req.Channel.Spec.ChannelInfo.Mattermost = &MattermostInfo{}
		for _, chName := range ruleManager.req.Channel.Notification.MatterMost.Channel {
			req.Channel.Spec.ChannelInfo.Mattermost.SendTo = append(req.Channel.Spec.ChannelInfo.Mattermost.SendTo, MattermostRecipient{
				MmChannel: chName,
			})
		}
	}

	if ruleManager.req.Channel.Notification.Email != nil {
		req.Channel.Spec.ChannelInfo.Email = &EmailInfo{}
		for _, user := range ruleManager.req.Channel.Recipient.Users {
			req.Channel.Spec.ChannelInfo.Email.SendTo = append(req.Channel.Spec.ChannelInfo.Email.SendTo, user)
		}
	}

	if ruleManager.req.Channel.Notification.Seatalk != nil {
		req.Channel.Spec.ChannelInfo.Seatalk = &SeatalkInfo{}
		for _, webHookUrl := range ruleManager.req.Channel.Notification.Seatalk.WebHook {
			req.Channel.Spec.ChannelInfo.Seatalk.SendTo = append(req.Channel.Spec.ChannelInfo.Seatalk.SendTo, SeatalkRecipient{
				BotWebhookUrl: webHookUrl,
			})
		}
	}

	// deal channel end

	// 构造 Web 回调
	if req.Channel.Spec.ChannelInfo.Webhook == nil {
		req.Channel.Spec.ChannelInfo.Webhook = &WebhookInfo{SendTo: []WebhookRecipient{}}
	}

	portalWebHook := WebhookRecipient{
		WebhookUrl: ChannelWebHookUrl,
		WebhookHeader: []WebhookHeader{{
			Key:   "Authorization",
			Value: "Bearer " + s.conf.CallBackToken,
		}},
		WebhookBody: NewWebhookBody(monitorRuleId, alertRuleUuid, channelUuid, channelName, ruleManager.req.Rule.AlertTemplateUUID, ruleManager.req.Rule.Name, ruleManager.req.Rule.GetAlertStrategy()),
		WebhookDesc: "slow query alert web hook.",
	}
	req.Channel.Spec.ChannelInfo.Webhook.AddWebhookRecipient(portalWebHook)
	return req
}

func (s *Service) genCreateStrategyRequest(ruleManager *RuleManager, monitorRuleId, channelUuid string) *CreateNotificationStrategyRequest {
	channelConfig := ChannelConfig{
		Names:  []string{getChannelName(channelUuid)},
		Config: &ChannelConfigDetail{},
	}

	if ruleManager.req.Channel.Notification.MatterMost != nil {
		channelConfig.Config.Mattermost = &MattermostConfig{
			ChannelConfigCommon: ChannelConfigCommon{
				RepeatInterval: ruleManager.req.Channel.Recipient.Interval,
				Enabled:        true,
			},
		}
	}

	if ruleManager.req.Channel.Notification.Email != nil {
		channelConfig.Config.Email = &EmailConfig{
			ChannelConfigCommon: ChannelConfigCommon{
				RepeatInterval:    ruleManager.req.Channel.Recipient.Interval,
				MessageTemplateId: s.conf.MonitorClientConfig.EmailMsgID,
				Enabled:           true,
			},
		}
	}

	if ruleManager.req.Channel.Notification.Seatalk != nil {
		channelConfig.Config.Seatalk = &SeatalkConfig{
			ChannelConfigCommon: ChannelConfigCommon{
				RepeatInterval:    ruleManager.req.Channel.Recipient.Interval,
				MessageTemplateId: s.conf.MonitorClientConfig.SeatalkMsgID,
				Enabled:           true,
			},
		}
	}

	channelConfig.Config.Webhook = &WebhookConfig{
		ChannelConfigCommon: ChannelConfigCommon{
			RepeatInterval: ruleManager.req.Channel.Recipient.Interval,
			Enabled:        true,
		},
	}

	req := &CreateNotificationStrategyRequest{
		BindLabelOrg: LabelOrg,
		ProjectName:  ProjectName,
		DisplayName:  fmt.Sprintf("strategy-%s-%s", monitorRuleId, channelUuid),
		BindFilter: [][]Bind{{{
			Key:      "m_rule_id",
			Operator: "=",
			Value:    monitorRuleId,
		}}},
		EnableResolvedMsg: &EnableResolvedMsg,
		IsDisabled:        0,
		ChannelConfig:     []ChannelConfig{channelConfig},
		StrategyType:      GENERAL,
	}
	return req
}

// GenUpdateRuleRequest 构建更新rule 请求结构，
func (s *Service) GenUpdateRuleRequest(remoteRule *GetSingleAlertRuleResponse, ruleManager *RuleManager, dataVersion string) *UpdateAlertRuleRequest {
	isDisabled := statusMap[strings.ToLower(ruleManager.req.Status)]
	if !strings.Contains(ruleManager.req.Rule.AlertMsg, fmt.Sprintf(link, url.QueryEscape(remoteRule.DisplayName))) {
		ruleManager.req.Rule.AlertMsg = ruleManager.req.Rule.AlertMsg + fmt.Sprintf(MessageLink, ruleManager.req.Rule.ExpressionValue, url.QueryEscape(ruleManager.req.Rule.Name))
	}
	if !strings.Contains(ruleManager.req.Rule.ResolveMsg, fmt.Sprintf(link, url.QueryEscape(remoteRule.DisplayName))) {
		ruleManager.req.Rule.ResolveMsg = ruleManager.req.Rule.ResolveMsg + fmt.Sprintf(MessageLink, ruleManager.req.Rule.ExpressionValue, url.QueryEscape(ruleManager.req.Rule.Name))
	}
	req := &UpdateAlertRuleRequest{
		AlertRule: Rule{
			MetricStoreNames: [][]string{{s.conf.MonitorClientConfig.MetricStoreNames}},
			AlertRuleContent: &RuleContent{
				Severity: ruleManager.req.Rule.Severity,
				Content: Content{
					For:                ruleManager.req.Rule.ForRange,
					EvaluationInterval: ruleManager.req.Rule.EvaluateEvery,
					Expr:               ruleManager.PromQL,
					MessageTemplate: MessageTemplate{
						FiringMessage:   ruleManager.req.Rule.AlertMsg,
						ResolvedMessage: ruleManager.req.Rule.ResolveMsg,
					},
				},
			},
			NotificationType: SstUpLater, // ruleManager.req.Rule.Trigger
			IsDisabled:       &isDisabled,
			DataVersion:      dataVersion,
			DisplayName:      ruleManager.req.Rule.Name,
		},
		UpdateMask: UpdateMask{[]string{"is_use_project_default_metric_store",
			"metric_store_names", "sop_ids", "alert_rule_content.severity",
			"alert_rule_content.content.evaluation_interval",
			"alert_rule_content.content.expr",
			"alert_rule_content.content.for",
			"alert_rule_content.content.message_template.firing_message",
			"alert_rule_content.content.message_template.resolved_message",
			"alert_rule_content.content.built_in_evaluation_interval",
			"display_name", "is_disabled", "notification_type",
			"grafana_url", "rule_desc",
			"alert_rule_labels"}},
	}
	return req
}

func (s *Service) GenUpdateStrategyRequest(ruleManager *RuleManager, channelUuid, dataVersion string, needKeepStrategyStatusIsDisabled bool) *UpdateNotificationStrategyRequest {
	channelConfig := ChannelConfig{
		Names:  []string{getChannelName(channelUuid)},
		Config: &ChannelConfigDetail{},
	}

	if ruleManager.req.Channel.Notification.MatterMost != nil {
		channelConfig.Config.Mattermost = &MattermostConfig{
			ChannelConfigCommon: ChannelConfigCommon{
				RepeatInterval: ruleManager.req.Channel.Recipient.Interval,
				Enabled:        true,
			},
		}
	}

	if ruleManager.req.Channel.Notification.Email != nil {
		channelConfig.Config.Email = &EmailConfig{
			ChannelConfigCommon: ChannelConfigCommon{
				RepeatInterval:    ruleManager.req.Channel.Recipient.Interval,
				MessageTemplateId: s.conf.MonitorClientConfig.EmailMsgID,
				Enabled:           true,
			},
		}
	}

	if ruleManager.req.Channel.Notification.Seatalk != nil {
		channelConfig.Config.Seatalk = &SeatalkConfig{
			ChannelConfigCommon: ChannelConfigCommon{
				RepeatInterval:    ruleManager.req.Channel.Recipient.Interval,
				MessageTemplateId: s.conf.MonitorClientConfig.SeatalkMsgID,
				Enabled:           true,
			},
		}
	}

	channelConfig.Config.Webhook = &WebhookConfig{
		ChannelConfigCommon: ChannelConfigCommon{
			RepeatInterval: ruleManager.req.Channel.Recipient.Interval,
			Enabled:        true,
		},
	}

	req := &UpdateNotificationStrategyRequest{
		AlertNotificationBind: NotificationBind{
			DataVersion:   dataVersion,
			ChannelConfig: []ChannelConfig{channelConfig},
		},
		UpdateMask: UpdateMask{[]string{"channel_config"}},
	}

	if needKeepStrategyStatusIsDisabled {
		isDisable := 0
		req.AlertNotificationBind.IsDisabled = &isDisable
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "is_disabled")
	}

	return req
}

// GenUpdateChannelRequest 构建更新channel 请求结构, 若是关闭rule状态，仅仅关闭rule状态，不关闭channel状态和strategy状态
func (s *Service) GenUpdateChannelRequest(ruleManager *RuleManager, monitorRuleId, alertRuleUuid, channelUuid, dataVersion string, needKeepStrategyStatusIsDisabled bool) *UpdateNotifyChannelRequest {
	req := &UpdateNotifyChannelRequest{
		NotifyChannel: NotifyChannel{
			DataVersion: dataVersion,
			Channel: Channel{
				Version: ChannelVersion,
				Spec:    ChannelSpec{ChannelInfo: ChannelInfo{}},
			},
		},
		UpdateMask: UpdateMask{[]string{"channel_config.spec.channel.tele.send_to", "channel_config.spec.channel.seatalk.send_to", "channel_config.spec.channel.email.send_to", "channel_config.spec.channel.mattermost.send_to", "channel_config.spec.channel.webhook.send_to", "channel_config.version", "display_name", "channel_config"}},
	}
	if needKeepStrategyStatusIsDisabled {
		isDisable := 0
		req.NotifyChannel.IsDisabled = &isDisable
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "is_disabled")
	}

	// deal channel start
	if ruleManager.req.Channel.Notification.MatterMost != nil {
		req.NotifyChannel.Channel.Spec.ChannelInfo.Mattermost = &MattermostInfo{}
		for _, chName := range ruleManager.req.Channel.Notification.MatterMost.Channel {
			req.NotifyChannel.Channel.Spec.ChannelInfo.Mattermost.SendTo = append(req.NotifyChannel.Channel.Spec.ChannelInfo.Mattermost.SendTo, MattermostRecipient{
				MmChannel: chName,
			})
		}
	}

	if ruleManager.req.Channel.Notification.Email != nil {
		req.NotifyChannel.Channel.Spec.ChannelInfo.Email = &EmailInfo{}
		for _, user := range ruleManager.req.Channel.Recipient.Users {
			req.NotifyChannel.Channel.Spec.ChannelInfo.Email.SendTo = append(req.NotifyChannel.Channel.Spec.ChannelInfo.Email.SendTo, user)
		}
	}

	if ruleManager.req.Channel.Notification.Seatalk != nil {
		req.NotifyChannel.Channel.Spec.ChannelInfo.Seatalk = &SeatalkInfo{}
		for _, webHookUrl := range ruleManager.req.Channel.Notification.Seatalk.WebHook {
			req.NotifyChannel.Channel.Spec.ChannelInfo.Seatalk.SendTo = append(req.NotifyChannel.Channel.Spec.ChannelInfo.Seatalk.SendTo, SeatalkRecipient{
				BotWebhookUrl: webHookUrl,
			})
		}
	}

	// deal channel end

	// 构造 Web 回调
	if req.NotifyChannel.Channel.Spec.ChannelInfo.Webhook == nil {
		req.NotifyChannel.Channel.Spec.ChannelInfo.Webhook = &WebhookInfo{SendTo: []WebhookRecipient{}}
	}

	portalWebHook := WebhookRecipient{
		WebhookUrl: ChannelWebHookUrl,
		WebhookHeader: []WebhookHeader{{
			Key:   "Authorization",
			Value: "Bearer " + s.conf.CallBackToken,
		}},
		WebhookBody: NewWebhookBody(monitorRuleId, alertRuleUuid, channelUuid, getChannelName(channelUuid), ruleManager.req.Rule.AlertTemplateUUID, ruleManager.req.Rule.Name, ruleManager.req.Rule.GetAlertStrategy()),
		WebhookDesc: "slow query alert web hook.",
	}
	req.NotifyChannel.Channel.Spec.ChannelInfo.Webhook.AddWebhookRecipient(portalWebHook)

	// 构造spec
	req.UpdateMask.Paths = append(req.UpdateMask.Paths, "channel_config.spec.channel")
	return req
}

func (s *Service) GetAlertTemplates() []*conf.AlertTemplate {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("GetAlertTemplates", sysMetrics.GetStatus(nil), time.Since(start))

	templates := make([]*conf.AlertTemplate, 0)
	for _, template := range s.conf.AlertTemplates {
		templates = append(templates, template)
	}
	return templates
}

func (s *Service) FindAlertTemplateByUUID(uuid string) (template *conf.AlertTemplate, err error) {
	var (
		ok    bool
		start = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("FindAlertTemplateByUUID", sysMetrics.GetStatus(nil), time.Since(start))

	if template, ok = s.conf.AlertTemplates[uuid]; !ok {
		return nil, fmt.Errorf("no template record for uuid:%s", uuid)
	}
	return template, nil
}

// NewWebhookBody 构建callback返回结构
func NewWebhookBody(alertRuleId, alertUuid, channelUuid, channelName, templateUUID, alertName, alertStrategy string) string {
	replacer := strings.NewReplacer("\n", "", "\r", "", "\t", " ")
	return replacer.Replace(fmt.Sprintf(hookBody, alertRuleId, alertUuid, channelUuid, channelName, templateUUID, alertName, alertStrategy))
}

func getChannelName(channelUuid string) string {
	return fmt.Sprintf("slow-query-channel-%s", channelUuid)
}
