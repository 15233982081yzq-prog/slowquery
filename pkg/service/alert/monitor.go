package alert

import (
	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/log"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	CreateAlterRulePath    = "/nodeapi/v1/alert_rules"
	GetSingleAlertRulePath = "/nodeapi/v1/alert_rules/%s"
	UpdateAlertRulePath    = "/nodeapi/v1/alert_rules/%s"
	DeleteAlertRulePath    = "/nodeapi/v1/alert_rules/%s"

	CreateNotifyChannelPath = "/nodeapi/v1/notify_channels"
	GetNotifyChannelPath    = "/nodeapi/v1/notify_channels/%s"
	UpdateNotifyChannelPath = "/nodeapi/v1/notify_channels/%s"
	DeleteNotifyChannelPath = "/nodeapi/v1/notify_channels/%s"

	CreateAlertRuleMuteConfigPath    = "/nodeapi/v1/alert_rule_mute_config"
	GetSingleAlertRuleMuteConfigPath = "/nodeapi/v1/alert_rule_mute_config/%s"
	UpdateAlertRuleMuteConfigPath    = "/nodeapi/v1/alert_rule_mute_config/%s"

	CreateAlertRuleACKConfigPath = "/nodeapi/v1/alert_events/%s:acknowledge"

	CreateNotificationStrategyPath    = "/nodeapi/v1/alert_notification_binds"
	GetSingleNotificationStrategyPath = "/nodeapi/v1/alert_notification_binds/%s"
	UpdateNotificationStrategyPath    = "/nodeapi/v1/alert_notification_binds/%s"
	DeleteNotificationStrategyPath    = "/nodeapi/v1/alert_notification_binds/%s"

	GetAlertEventPath       = "/event-distribute/openapi/v2/events?auth_project=%s&auth_event_set_id=%s&start_time=%d&end_time=%d&project_name=%s&page=1&size=1000&annotations={database_name:%s}"
	GetAlertEventDetailPath = "/event-distribute/openapi/v2/event/%s?auth_project=%s&auth_event_set_id=%s"
)

const (
	ChannelVersion    = "mon.notifyChannel.channelConfig/v1"
	ChannelDesc       = "Auto create by slow-query, alert_rule_uuid=%s, channel_uuid=%s"
	ChannelWebHookUrl = "https://space.shopee.io/rds/smart/v1/alert_api/live/slowquery/monitor/alert_callback"
)

const (
	LabelOrg    = "engineering_infra.infra_products"
	ProjectName = "shopee.engineering_infra.infra_products.db_products.slowquery"
)

const (
	GENERAL = "GENERAL"
)

const (
	RuleType = "Advanced"
)

const (
	DataVersionDelta = 0
)

const (
	RuleDesc = "Auto create slow query, alert_rule_uuid=%s, channel_uuid=%s"
)

const (
	Source = "slow-query"
)

var (
	MonitorNotFoundErr = errors.New("monitor not fount record, you find has be deleted")
)

type RuleManager struct {
	req    *request.AlertCommonRequest
	PromQL string
}

func (r *RuleManager) ChannelTypes() (types []string) {
	if r.req.Channel.Notification.MatterMost != nil {
		types = append(types, "MatterMost")
	}
	if r.req.Channel.Notification.Seatalk != nil {
		types = append(types, "Seatalk")
	}
	if r.req.Channel.Notification.Email != nil {
		types = append(types, "Email")
	}
	if r.req.Channel.Notification.Phone != nil {
		types = append(types, "Phone")
	}
	return
}

type MonitorClient struct {
	client       *http.Client
	baseURL      string
	accessToken  string
	tokenLock    sync.RWMutex
	clientId     string
	clientSecret string
}

type FetchTokenRequest struct {
	ClientId     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

type HttpErrorResponse struct {
	Code    int       `json:"code,omitempty"`
	Message string    `json:"message,omitempty"`
	Details []Details `json:"details"`
}

type Details struct {
	FieldViolations []struct {
		Field       string `json:"field"`
		Description string `json:"description"`
	} `json:"field_violations"`
}

type EmptyBody struct{}

func NewMonitorClient(conf *conf.MonitorClientConfig) (*MonitorClient, error) {
	httpClient := &http.Client{
		Timeout: conf.Timeout * time.Second, // 设置HTTP客户端超时时间
	}

	mc := &MonitorClient{
		client:       httpClient,
		baseURL:      conf.BaseURL,
		clientId:     conf.ClientID,
		clientSecret: conf.ClientSecret,
	}

	// 初始化时获取Token
	if err := mc.FetchToken(); err != nil {
		log.Warningf("Fetch monitor token err:%v", err)
		return nil, err
	}

	// 启动异步协程定时更新token
	go mc.startTokenUpdater()

	return mc, nil

}

// startTokenUpdater 启动异步协程定时更新token
func (mc *MonitorClient) startTokenUpdater() {
	// 计算下一个凌晨00:00的时间点
	now := time.Now()
	nextMidnight := now.Add(time.Hour * 24)
	nextMidnight = time.Date(nextMidnight.Year(), nextMidnight.Month(), nextMidnight.Day(), 0, 0, 0, 0, nextMidnight.Location())

	// 计算初始等待时间
	initialWait := nextMidnight.Sub(now)

	// 定时更新token
	time.Sleep(initialWait)
	ticker := time.NewTicker(time.Hour * 24)
	for {
		select {
		case <-ticker.C:
			err := mc.FetchToken()
			log.Warningf("Fetch monitor token err:%v", err)
		}
	}
}

// FetchToken 用于获取和更新Token
func (mc *MonitorClient) FetchToken() error {
	var (
		payloadBytes []byte
		req          *http.Request
		resp         *http.Response
		err          error
	)
	data := FetchTokenRequest{
		ClientId:     mc.clientId,
		ClientSecret: mc.clientSecret,
	}

	url := mc.baseURL + "/nodeapi/v1/open_api/token"
	if payloadBytes, err = json.Marshal(data); err != nil {
		return err
	}

	if req, err = http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes)); err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if resp, err = mc.client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	// todo get token
	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return fmt.Errorf("failed to get access token")
	}

	// 更新 token
	mc.setToken(accessToken)
	return nil
}

func (mc *MonitorClient) getToken() string {
	mc.tokenLock.RLock()
	defer mc.tokenLock.RUnlock()
	token := mc.accessToken
	return token
}

// privateHTTPCall 是一个私有方法，用于抽象公共的HTTP请求逻辑
func (mc *MonitorClient) privateHTTPCall(method, path string, data interface{}) (*http.Response, error) {
	var (
		payload []byte
		req     *http.Request
		err     error
	)
	token := mc.getToken()

	url := mc.baseURL + path
	if payload, err = json.Marshal(data); err != nil {
		return nil, err
	}

	if req, err = http.NewRequest(method, url, bytes.NewBuffer(payload)); err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return mc.client.Do(req)
}

func (mc *MonitorClient) doHttp(path string, method string, req interface{}, resp interface{}) error {
	var (
		err      error
		body     []byte
		httpResp *http.Response
	)

	if httpResp, err = mc.privateHTTPCall(method, path, req); err != nil {
		log.Errorf("http call error, path:%s, err:%s", path, err.Error())
		resp = nil
		return err
	}

	defer httpResp.Body.Close()

	if body, err = ioutil.ReadAll(httpResp.Body); err != nil {
		log.Errorf("read body error:%v", err)
		resp = nil
		return err
	}

	if !isHttpOK(httpResp.StatusCode) {
		if httpResp.StatusCode == http.StatusNotFound {
			return MonitorNotFoundErr
		}
		errResp := &HttpErrorResponse{}
		if err = json.Unmarshal(body, errResp); err != nil {
			log.Errorf("json parse failed, %v", err)
			resp = nil
			return err
		}
		log.Errorf("http return not ok, path:%s, res:%s", path, string(body))
		resp = nil
		detail, _ := json.Marshal(errResp.Details)
		return fmt.Errorf("http err code:%v, msg:%s, detail:%s", errResp.Code, errResp.Message, string(detail))
	}

	if httpResp.StatusCode == http.StatusNoContent {
		return nil
	}

	if method != http.MethodDelete {
		if err = json.Unmarshal(body, resp); err != nil {
			log.Errorf("json parse failed, err:%v", err)
			resp = nil
			return err
		}
	}

	return nil
}

func (mc *MonitorClient) setToken(accessToken string) {
	mc.tokenLock.Lock()
	defer mc.tokenLock.Unlock()
	mc.accessToken = accessToken
}

func isHttpOK(StatusCode int) bool {
	return StatusCode >= 200 && StatusCode < 300
}

func (mc *MonitorClient) CreateNotifyChannel(req *CreateNotifyChannelRequest) (*CreateNotifyChannelResponse, error) {
	var (
		err error
	)
	resp := &CreateNotifyChannelResponse{}
	if err = mc.doHttp(CreateNotifyChannelPath, http.MethodPost, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) CreateAlterRule(req *CreateAlertRuleRequest) (*CreateAlertRuleResponse, error) {
	var (
		err  error
		resp = &CreateAlertRuleResponse{}
	)

	if err = mc.doHttp(CreateAlterRulePath, http.MethodPost, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) CreateNotificationStrategy(req *CreateNotificationStrategyRequest) (*CreateNotificationStrategyResponse, error) {
	var (
		err  error
		resp = &CreateNotificationStrategyResponse{}
	)

	if err = mc.doHttp(CreateNotificationStrategyPath, http.MethodPost, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) GetSingleNotificationStrategy(strategyId string) (*GetSingleNotificationStrategyResponse, error) {
	var (
		err  error
		resp = &GetSingleNotificationStrategyResponse{}
	)
	if err = mc.doHttp(fmt.Sprintf(GetSingleNotificationStrategyPath, strategyId), http.MethodGet, EmptyBody{}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) GetSingleAlertRule(ruleId string) (*GetSingleAlertRuleResponse, error) {
	var (
		err  error
		resp = &GetSingleAlertRuleResponse{}
	)
	if err = mc.doHttp(fmt.Sprintf(GetSingleAlertRulePath, ruleId), http.MethodGet, EmptyBody{}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) GetNotifyChannel(ruleId string) (*GetSingleNotifyChannelResponse, error) {
	var (
		err  error
		resp = &GetSingleNotifyChannelResponse{}
	)
	if err = mc.doHttp(fmt.Sprintf(GetNotifyChannelPath, ruleId), http.MethodGet, EmptyBody{}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) UpdateChannel(channelName string, req *UpdateNotifyChannelRequest) (resp *UpdateNotifyChannelResponse, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("MonitorClient.UpdateChannel", sysMetrics.GetStatus(err), time.Since(start))
	}()
	resp = &UpdateNotifyChannelResponse{}
	if err = mc.doHttp(fmt.Sprintf(UpdateNotifyChannelPath, channelName), http.MethodPatch, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) UpdateStrategy(strategyId string, req *UpdateNotificationStrategyRequest) (resp *UpdateNotificationStrategyResponse, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("MonitorClient.UpdateStrategy", sysMetrics.GetStatus(err), time.Since(start))
	}()
	resp = &UpdateNotificationStrategyResponse{}
	if err = mc.doHttp(fmt.Sprintf(UpdateNotificationStrategyPath, strategyId), http.MethodPatch, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) UpdateAlertRule(ruleId string, req *UpdateAlertRuleRequest) (resp *UpdateAlertRuleResponse, err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("MonitorClient.UpdateAlertRule", sysMetrics.GetStatus(err), time.Since(start))
	}()
	resp = &UpdateAlertRuleResponse{}
	if err = mc.doHttp(fmt.Sprintf(UpdateAlertRulePath, ruleId), http.MethodPatch, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) DeleteAlertRule(ruleId string) error {
	return mc.doHttp(fmt.Sprintf(DeleteAlertRulePath, ruleId), http.MethodDelete, EmptyBody{}, EmptyBody{})
}

func (mc *MonitorClient) DeleteNotificationStrategy(strategyId string) error {
	return mc.doHttp(fmt.Sprintf(DeleteNotificationStrategyPath, strategyId), http.MethodDelete, &DeleteNotificationStrategyRequest{ProjectName: ProjectName}, EmptyBody{})
}

func (mc *MonitorClient) DeleteAlertChannel(channelName string) error {
	return mc.doHttp(fmt.Sprintf(DeleteNotifyChannelPath, channelName), http.MethodDelete, &DeleteNotifyChannelRequest{ProjectName: ProjectName}, EmptyBody{})
}

func (mc *MonitorClient) CreateAlertRuleMuteConfig(req *CreateAlertRuleMuteConfigRequest) (*CreateAlertRuleMuteConfigResponse, error) {
	var (
		err  error
		resp = &CreateAlertRuleMuteConfigResponse{}
	)
	if err = mc.doHttp(CreateAlertRuleMuteConfigPath, http.MethodPost, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) GetAlertEvent(req *GetAlertEventRequest) (*GetAlertEventResponse, error) {
	var (
		err  error
		resp = &GetAlertEventResponse{}
	)
	url := fmt.Sprintf(GetAlertEventPath, req.AuthProject, req.AuthEventSetId, req.StartTime, req.EndTime, req.ProjectName, req.Annotations["database_name"])
	if err = mc.doHttp(url, http.MethodGet, EmptyBody{}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) GetAlertEventDetailByAlertId(alertId, authProject, authEventSetId string) (*GetAlertEventDetailResponse, error) {
	var (
		err  error
		resp = &GetAlertEventDetailResponse{}
	)
	url := fmt.Sprintf(GetAlertEventDetailPath, alertId, authProject, authEventSetId)
	if err = mc.doHttp(url, http.MethodGet, EmptyBody{}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) UpdateSingleAlertRuleMuteConfig(ruleId string, req *UpdateAlertRuleMuteConfigRequest) (*UpdateAlertRuleMuteConfigResponse, error) {
	var (
		err  error
		resp = &UpdateAlertRuleMuteConfigResponse{}
	)
	if err = mc.doHttp(fmt.Sprintf(UpdateAlertRuleMuteConfigPath, ruleId), http.MethodPatch, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (mc *MonitorClient) CreateAck(ruleId string) error {
	return mc.doHttp(fmt.Sprintf(CreateAlertRuleACKConfigPath, ruleId), http.MethodPost, EmptyBody{}, EmptyBody{})
}

func (mc *MonitorClient) GetSingleAlertRuleMuteConfig(ruleId string) (*GetSingleAlertRuleMuteConfigResponse, error) {
	var (
		err  error
		resp = &GetSingleAlertRuleMuteConfigResponse{}
	)
	if err = mc.doHttp(fmt.Sprintf(GetSingleAlertRuleMuteConfigPath, ruleId), http.MethodGet, EmptyBody{}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
