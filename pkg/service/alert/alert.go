package alert

import (
	alertConf "smart-slowquery/conf/alert"
	stringUtil "smart-slowquery/internal/util/string"
	timeUtil "smart-slowquery/internal/util/time"
	httpReq "smart-slowquery/pkg/http/request"
	responseMeta "smart-slowquery/pkg/http/response"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	storeMsql "smart-slowquery/pkg/store/mysql"
	storeReq "smart-slowquery/pkg/store/request"

	"smart-slowquery/conf"
	"smart-slowquery/internal/model/alert"
	"smart-slowquery/internal/util/function"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/oplog"
	"smart-slowquery/pkg/service/uic"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/response"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	splitMessageKey = "?name="
)

type duringUnmuteMessage struct {
	m       map[string]response.AlertMessage
	mux     *sync.RWMutex
	AlertId string
	EndTime time.Time
}

type Service struct {
	conf                *alertConf.Config
	db                  storeMsql.DB
	monitorClient       *MonitorClient
	dodClient           *DodClient
	adminList           []string
	httpClient          *http.Client
	ck                  store.CKStore
	alertOpLog          *oplog.AlertOpLog
	duringUnmuteMessage *duringUnmuteMessage
}

func NewAlertService(conf *alertConf.Config, mysqlDB storeMsql.DB, ck store.CKStore, alertOpLog *oplog.AlertOpLog) (*Service, error) {
	var (
		monitorClient *MonitorClient
		dodClient     *DodClient
		err           error
	)
	if monitorClient, err = NewMonitorClient(conf.MonitorClientConfig); err != nil {
		return nil, err
	}
	if dodClient, err = NewDodClient(conf.DodConfig); err != nil {
		return nil, err
	}
	service := &Service{
		conf:          conf,
		db:            mysqlDB,
		monitorClient: monitorClient,
		dodClient:     dodClient,
		adminList:     []string{"hanwen.liu@shopee.com", "jian.bian@shopee.com"},
		ck:            ck,
		httpClient: &http.Client{
			Timeout: timeUtil.TenSecond, // 设置HTTP客户端超时时间
		},
		alertOpLog:          alertOpLog,
		duringUnmuteMessage: &duringUnmuteMessage{},
	}
	go function.Loop("mute_consistency", func() error {
		if err := service.alwaysUpdateMuteDBStatus(); err != nil {
			return err
		}
		return service.accordMutedUpdateStatus()
	}, timeUtil.TenSecond)

	return service, nil
}

func (s *Service) CreateMonitorRule(ruleManagers []*RuleManager) (alertRuleUUIDList, channelUUIDList []string, err error) {
	var (
		actionType                                                                = oplog.CREATE_ALERT
		createChannelRespList                                                     = make([]*CreateNotifyChannelResponse, len(ruleManagers))
		createAlertRuleRespList                                                   = make([]*CreateAlertRuleResponse, len(ruleManagers))
		createStrategyRespList                                                    = make([]*CreateNotificationStrategyResponse, len(ruleManagers))
		ruleInfoList                                                              = make([]*alert.RuleTab, len(ruleManagers))
		channelInfoList                                                           = make([]*alert.ChannelTab, len(ruleManagers))
		strategyInfoList                                                          = make([]*alert.StrategyTab, len(ruleManagers))
		dbListJsonList, userListJsonList, channelConfigJsonList, strategyMetaJson = make([][]byte, len(ruleManagers)), make([][]byte, len(ruleManagers)), make([][]byte, len(ruleManagers)), make([][]byte, len(ruleManagers))
		monitorIds, monitorChannelNames, monitorStrategyIds                       []string
		start                                                                     = time.Now()
	)

	defer func() {
		if err != nil {
			log.Errorf("alertSrv.CreateMonitorRule error:%s,deleteMonitorSet monitorIds:%v,monitorChannelNames:%v", err.Error(), monitorIds, monitorChannelNames)
			if e := s.DeleteAlertRulesAndChannelsAndStrategy(monitorIds, monitorChannelNames, monitorStrategyIds); e != nil {
				log.Errorf("CreateMonitorRule error,do DeleteAlertRulesAndChannels failed:%s monitorIds:%v,monitorChannelNames:%v", e.Error(), monitorIds, monitorChannelNames)
			}
		}
		sysMetrics.CollectServiceMetrics("alertSrv.CreateMonitorRule", sysMetrics.GetStatus(err), time.Since(start))
	}()

	for i, ruleManager := range ruleManagers {
		// gen uuid
		alertRuleUUID := uuid.New().String()
		channelUUID := uuid.New().String()

		// 1. go monitor create alert rule
		if createAlertRuleRespList[i], err = s.monitorClient.CreateAlterRule(s.genCreateAlertRuleRequest(ruleManager, alertRuleUUID, channelUUID)); err != nil {
			log.Warningf("alertSrv.CreateMonitorRule create alert rule failed:%v", err)
			return nil, nil, err
		}
		log.Infof("alertSrv.CreateMonitorRule create alert rule success, alert_uuid=%s", alertRuleUUID)
		monitorIds = append(monitorIds, createAlertRuleRespList[i].RuleId)

		// 2. go monitor create channel
		if createChannelRespList[i], err = s.monitorClient.CreateNotifyChannel(s.genCreateChannelRequest(ruleManager, createAlertRuleRespList[i].RuleId, alertRuleUUID, channelUUID)); err != nil {
			log.Warningf("alertSrv.CreateMonitorRule create channel failed:%v", err)
			return nil, nil, err
		}
		log.Infof("alertSrv.CreateMonitorRule create channel success, channel_uuid=%s", channelUUID)
		monitorChannelNames = append(monitorChannelNames, createChannelRespList[i].ChannelName)

		// 3. go monitor create strategy, bind rule and channel
		if createStrategyRespList[i], err = s.monitorClient.CreateNotificationStrategy(s.genCreateStrategyRequest(ruleManager, createAlertRuleRespList[i].RuleId, channelUUID)); err != nil {
			log.Warningf("alertSrv.CreateMonitorRule create alert strategy failed:%v", err)
			return nil, nil, err
		}
		log.Infof("alertSrv.CreateMonitorRule create strategy success, alert_uuid=%s", alertRuleUUID)
		monitorStrategyIds = append(monitorStrategyIds, createStrategyRespList[i].BindId)

		if dbListJsonList[i], err = json.Marshal(ruleManager.req.DBS); err != nil {
			log.Errorf("alertSrv.CreateMonitorRule service CreateMonitorRule Marshal dbListJson error:%s", err.Error())
			return nil, nil, err
		}

		if userListJsonList[i], err = json.Marshal(ruleManager.req.Channel.Recipient.Users); err != nil {
			log.Errorf("alertSrv.CreateMonitorRule service CreateMonitorRule Marshal userListJson error:%s", err.Error())
			return nil, nil, err
		}

		if channelConfigJsonList[i], err = json.Marshal(ruleManager.req.Channel.Notification); err != nil {
			log.Errorf("alertSrv.CreateMonitorRule service CreateMonitorRule Marshal channelConfigJson error:%s", err.Error())
			return nil, nil, err
		}

		if strategyMetaJson[i], err = json.Marshal(createStrategyRespList[i].BindFilter); err != nil {
			log.Errorf("alertSrv.CreateMonitorRule service CreateMonitorRule Marshal channelConfigJson error:%s", err.Error())
			return nil, nil, err
		}

		ruleInfoList[i] = &alert.RuleTab{
			AlertRuleUUID:      alertRuleUUID,
			ChannelUUID:        channelUUID,
			StrategyID:         createStrategyRespList[i].BindId,
			CMDB:               ruleManager.req.CMDB,
			DBEnv:              ruleManager.req.DBEnv,
			Trigger:            ruleManager.req.Rule.Trigger,
			MonitorRuleID:      createAlertRuleRespList[i].RuleId,
			AlertDisplayName:   ruleManager.req.Rule.Name,
			TemplateName:       s.conf.AlertTemplates[ruleManager.req.Rule.AlertTemplateUUID].Name,
			Severity:           ruleManager.req.Rule.Severity,
			PromQL:             ruleManager.PromQL,
			Expression:         ruleManager.req.Rule.Expression,
			ExpressionValue:    ruleManager.req.Rule.ExpressionValue,
			ForRange:           ruleManager.req.Rule.ForRange,
			EvaluationInterval: ruleManager.req.Rule.EvaluateEvery,
			AlarmMsg:           ruleManager.req.Rule.AlertMsg,
			ResolveMsg:         ruleManager.req.Rule.ResolveMsg,
			SoftStatus:         alert.NormalSoftStatus,
			Status:             ruleManager.req.Status,
			DbsJson:            string(dbListJsonList[i]),
			Creator:            ruleManager.req.Applicant,
			ChannelType:        strings.Join(ruleManager.ChannelTypes(), "/"),
			CreateTime:         time.Now(),
			UpdateTime:         time.Now(),
		}
		channelInfoList[i] = &alert.ChannelTab{
			ChannelUUID: channelUUID,
			ChannelName: createChannelRespList[i].ChannelName,
			DodID:       ruleManager.req.Channel.Recipient.Dod,
			UsersJson:   string(userListJsonList[i]),
			Interval:    ruleManager.req.Channel.Recipient.Interval,
			MetaJson:    string(channelConfigJsonList[i]),
			Status:      ruleManager.req.Status,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}
		strategyInfoList[i] = &alert.StrategyTab{
			StrategyName:   createStrategyRespList[i].DisplayName,
			StrategyID:     createStrategyRespList[i].BindId,
			MetaJson:       string(strategyMetaJson[i]),
			StrategyStatus: ruleManager.req.Status,
			CreateTime:     time.Now(),
			UpdateTime:     time.Now(),
		}
		alertRuleUUIDList = append(alertRuleUUIDList, alertRuleUUID)
		channelUUIDList = append(channelUUIDList, channelUUID)
	}

	// 都创建成功，开始入库

	if err = alert.SaveRulesAndChannelsAndStrategy(s.db, ruleInfoList, channelInfoList, strategyInfoList); err != nil {
		log.Errorf("alertSrv.CreateMonitorRule service CreateMonitorRule SaveRulesAndChannelsAndStrategy error:%s", err.Error())
		return nil, nil, err
	}

	// record operator log
	if len(ruleManagers) > 1 {
		actionType = oplog.BATCH_CREATE_ALERT
	}

	if err = s.alertOpLog.Record(ruleManagers[0].req.Applicant, strings.Join(alertRuleUUIDList, ","), oplog.CREATE, actionType, ruleManagers[0].req.DBEnv, "", map[string]interface{}{
		"rule":     ruleInfoList,
		"channel":  channelInfoList,
		"strategy": strategyInfoList,
	}); err != nil {
		log.Warningf("alertSrv.CreateMonitorRule CreateMonitorRule alertOpLog failed error:%s", err.Error())
	}

	return alertRuleUUIDList, channelUUIDList, nil
}

// UpdateMonitorRule ps:更新状态的情况下， 只有rule更新状态， channel和strategy不能更新状态，为了接受close消息
func (s *Service) UpdateMonitorRule(ruleManager *RuleManager, alertRuleUUID string) (ruleUUID, channelUUID string, err error) {
	var (
		rule, ruleInfo                                                *alert.RuleTab
		channel, channelInfo                                          *alert.ChannelTab
		strategy, strategyInfo                                        *alert.StrategyTab
		ruleVersion, channelVersion, strategyVersion                  string
		remoteRule                                                    *GetSingleAlertRuleResponse
		remoteChannel                                                 *GetSingleNotifyChannelResponse
		remoteStrategy                                                *GetSingleNotificationStrategyResponse
		alertRuleResp                                                 *UpdateAlertRuleResponse
		channelRuleResp                                               *UpdateNotifyChannelResponse
		strategyResp                                                  *UpdateNotificationStrategyResponse
		dbListJson, userListJson, channelConfigJson, strategyMetaJson []byte
	)

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("alertSrv.UpdateMonitorRule", sysMetrics.GetStatus(err), time.Since(start))

	if rule, channel, strategy, err = s.getLocalAlertRule(alertRuleUUID); err != nil {
		return "", "", fmt.Errorf("getLocalAlertRule error:%s", err.Error())
	}
	ruleUUID = rule.AlertRuleUUID
	channelUUID = rule.ChannelUUID

	if remoteRule, remoteChannel, remoteStrategy, err = s.getRemoteAlertRule(rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID); err != nil {
		log.Errorf("alertSrv.UpdateMonitorRule UpdateMonitorRule monitorRuleID:%s,channelName:%s,strategyID:%s,error:%s", rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID, err.Error())
		return
	}

	if err = s.syncStatusFromMonitor(rule, channel, strategy, remoteRule, remoteChannel, remoteStrategy); err != nil {
		log.Errorf("alertSrv.UpdateMonitorRule syncStatusFromMonitor monitorRuleID:%s,channelName:%s,strategyID:%s,error:%s", rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID, err.Error())
		return
	}

	if ruleVersion, err = stringUtil.IncrementString(remoteRule.DataVersion, DataVersionDelta); err != nil {
		log.Errorf("service UpdateMonitorRule incrementString ruleVersion error:%s", err.Error())
		return
	}

	if channelVersion, err = stringUtil.IncrementString(remoteChannel.DataVersion, DataVersionDelta); err != nil {
		log.Errorf("service UpdateMonitorRule incrementString channelVersion error:%s", err.Error())
		return
	}

	if strategyVersion, err = stringUtil.IncrementString(remoteStrategy.DataVersion, DataVersionDelta); err != nil {
		log.Errorf("service UpdateMonitorRule incrementString strategyVersion error:%s", err.Error())
		return
	}

	// 3.update remote， if remote channel or strategy has been deleted,need recreate
	if alertRuleResp, err = s.monitorClient.UpdateAlertRule(rule.MonitorRuleID, s.GenUpdateRuleRequest(remoteRule, ruleManager, ruleVersion)); err != nil {
		log.Errorf("service UpdateMonitorRule UpdateAlertRule error:%s", err.Error())
		return
	}

	if channelRuleResp, err = s.monitorClient.UpdateChannel(channel.ChannelName, s.GenUpdateChannelRequest(ruleManager, alertRuleResp.RuleId, alertRuleUUID, channelUUID, channelVersion, remoteChannel.Disable())); err != nil {
		log.Errorf("service UpdateMonitorRule UpdateChannel error:%s", err.Error())
		return
	}

	if strategyResp, err = s.monitorClient.UpdateStrategy(remoteStrategy.BindId, s.GenUpdateStrategyRequest(ruleManager, channel.ChannelUUID, strategyVersion, remoteStrategy.Disable())); err != nil {
		log.Errorf("service UpdateMonitorRule UpdateStrategy error:%s", err.Error())
		return
	}

	// 4.update local
	if dbListJson, err = json.Marshal(ruleManager.req.DBS); err != nil {
		log.Errorf("service UpdateMonitorRule dbListJson Marshal error:%s", err.Error())
		return
	}

	if userListJson, err = json.Marshal(ruleManager.req.Channel.Recipient.Users); err != nil {
		log.Errorf("service UpdateMonitorRule userListJson Marshal error:%s", err.Error())
		return
	}

	if channelConfigJson, err = json.Marshal(ruleManager.req.Channel); err != nil {
		log.Errorf("service UpdateMonitorRule channelConfigJson Marshal error:%s", err.Error())
		return
	}

	if strategyMetaJson, err = json.Marshal(strategyResp.BindFilter); err != nil {
		log.Errorf("service UpdateMonitorRule channelConfigJson Marshal error:%s", err.Error())
		return
	}

	ruleInfo = &alert.RuleTab{
		AlertRuleUUID:      alertRuleUUID,
		ChannelUUID:        channelUUID,
		StrategyID:         strategyResp.BindId,
		CMDB:               ruleManager.req.CMDB,
		DBEnv:              ruleManager.req.DBEnv,
		Trigger:            ruleManager.req.Rule.Trigger,
		MonitorRuleID:      alertRuleResp.RuleId,
		AlertDisplayName:   ruleManager.req.Rule.Name,
		TemplateName:       s.conf.AlertTemplates[ruleManager.req.Rule.AlertTemplateUUID].Name,
		Severity:           ruleManager.req.Rule.Severity,
		PromQL:             ruleManager.PromQL,
		Expression:         ruleManager.req.Rule.Expression,
		ExpressionValue:    ruleManager.req.Rule.ExpressionValue,
		ForRange:           ruleManager.req.Rule.ForRange,
		EvaluationInterval: ruleManager.req.Rule.EvaluateEvery,
		AlarmMsg:           ruleManager.req.Rule.AlertMsg,
		ResolveMsg:         ruleManager.req.Rule.ResolveMsg,
		Status:             ruleManager.req.Status,
		SoftStatus:         rule.SoftStatus, // 更新不修改软状态
		DbsJson:            string(dbListJson),
		Modifier:           ruleManager.req.Applicant,
		ChannelType:        strings.Join(ruleManager.ChannelTypes(), "/"),
		CreateTime:         rule.CreateTime,
		UpdateTime:         time.Now(),
	}

	channelInfo = &alert.ChannelTab{
		ChannelUUID: channelUUID,
		ChannelName: channelRuleResp.ChannelName,
		DodID:       ruleManager.req.Channel.Recipient.Dod,
		UsersJson:   string(userListJson),
		Interval:    ruleManager.req.Channel.Recipient.Interval,
		MetaJson:    string(channelConfigJson),
		Status:      httpReq.ENABLE, // 使用ENABLE状态
		CreateTime:  channel.CreateTime,
		UpdateTime:  time.Now(),
	}

	strategyInfo = &alert.StrategyTab{
		StrategyName:   strategyResp.DisplayName,
		StrategyID:     strategyResp.BindId,
		MetaJson:       string(strategyMetaJson),
		StrategyStatus: httpReq.ENABLE, // 使用ENABLE状态
		CreateTime:     strategy.CreateTime,
		UpdateTime:     time.Now(),
	}

	if err = alert.UpdateRuleAndChannelAndStrategy(s.db, ruleInfo, channelInfo, strategyInfo); err != nil {
		log.Errorf("alertSrv.UpdateMonitorRule UpdateRuleAndChannelAndStrategy rule:%v,channel:%v,strategyInfo:%v ,error:%v", ruleInfo, channelInfo, strategyInfo, err.Error())
		return
	}

	// record operator log
	if err = s.alertOpLog.Record(ruleManager.req.Applicant, alertRuleUUID, oplog.UPDATE, oplog.UPDATE_ALERT, ruleManager.req.DBEnv, map[string]interface{}{
		"rule":     rule,
		"channel":  channel,
		"strategy": strategy,
	}, map[string]interface{}{
		"rule":     ruleInfo,
		"channel":  channelInfo,
		"strategy": strategyInfo,
	}); err != nil {
		log.Warningf("alertSrv.UpdateMonitorRule alertOpLog failed:%s", err.Error())
	}

	return
}

func (s *Service) ChangeRuleStatus(alertRuleUUID, status, operator string) (oldStatus, alertName string, err error) {
	var (
		rule     *alert.RuleTab
		channel  *alert.ChannelTab
		strategy *alert.StrategyTab
	)

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("alertSrv.ChangeRuleStatus", sysMetrics.GetStatus(err), time.Since(start))

	if rule, channel, strategy, err = s.getLocalAlertRule(alertRuleUUID); err != nil {
		log.Errorf("alertSrv.ChangeRuleStatus getLocalAlertRule alertRuleUUID:%s,error:%s", alertRuleUUID, err.Error())
		return "", "", fmt.Errorf("getLocalAlertRule ruleUUID:%s,error:%s", alertRuleUUID, err.Error())
	}
	oldStatus = rule.Status

	switch strings.ToLower(status) {
	case httpReq.DISABLE:
		if err = s.DisableAlertRuleAndChannel(rule, channel, strategy); err != nil {
			log.Errorf("alertSrv.ChangeRuleStatus DisableAlertRuleAndChannel error:%s", err.Error())
			return "", "", err
		}
	case httpReq.ENABLE:
		if err = s.EnableAlertRuleAndChannel(rule, channel, strategy); err != nil {
			log.Errorf("alertSrv.ChangeRuleStatus EnableAlertRuleAndChannel error:%s", err.Error())
			return "", "", err
		}
	default:
		return "", "", fmt.Errorf("not match status:%s, uuid:%s", status, alertRuleUUID)
	}

	rule.Status, channel.Status = status, status
	rule.Modifier = operator
	// 只更新rule的状态，并不更新channel和strategy的状态
	if err = alert.UpdateRuleStatus(s.db, rule); err != nil {
		log.Errorf("alertSrv.ChangeRuleStatus UpdateRuleStatus rule:%v,error:%s", rule, err.Error())
		return "", "", err
	}

	return oldStatus, rule.AlertDisplayName, nil
}

func (s *Service) DeleteRules(alertRuleUUID string) (oldRule *alert.RuleTab, err error) {
	var (
		rule     *alert.RuleTab
		channel  *alert.ChannelTab
		strategy *alert.StrategyTab
	)

	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.DeleteRules", sysMetrics.GetStatus(err), time.Since(start))

	if rule, err = s.FindAlertRuleFromDBByUUID(alertRuleUUID); err != nil {
		log.Errorf("alertSrv.DeleteRules FindAlertRuleFromDBByUUID alertRuleUUID:%s,error:%s", alertRuleUUID, err.Error())
		return nil, err
	}

	if channel, err = s.FindAlertChannelFromDBByUUID(rule.ChannelUUID); err != nil {
		log.Errorf("alertSrv.DeleteRules FindAlertChannelFromDBByUUID channelUUID:%s ,error:%s", rule.ChannelUUID, err.Error())
		return nil, err
	}

	if strategy, err = s.FindStrategyFromDBByID(rule.StrategyID); err != nil {
		log.Errorf("alertSrv.DeleteRules FindStrategyFromDBByID strategyID:%s ,error:%s", rule.StrategyID, err.Error())
		return nil, err
	}

	// 直接删除远端-检查远端monitor的相关状态
	if err = s.monitorClient.DeleteAlertRule(rule.MonitorRuleID); err != nil {
		log.Warningf("alertSrv.DeleteRules DeleteAlertRule monitorRuleID:%s, error:%s", rule.MonitorRuleID, err.Error())
	}

	if err = s.monitorClient.DeleteNotificationStrategy(rule.StrategyID); err != nil {
		log.Warningf("alertSrv.DeleteRules DeleteNotificationStrategy StrategyID:%s, error:%s", rule.StrategyID, err.Error())
	}

	if err = s.monitorClient.DeleteAlertChannel(channel.ChannelName); err != nil {
		log.Warningf("alertSrv.DeleteRules DeleteAlertChannel channelName:%s, error:%s", channel.ChannelName, err.Error())
	}

	// 数据库软状态置为delete，等收到close消息时候 再进行删除
	if err = alert.DeleteRuleAbout(s.db, rule, channel, strategy); err != nil {
		log.Warningf("alertSrv.DeleteRules DeleteRuleAbout  monitorRuleID:%s,channelName:%s,StrategyID:%s,error:%s", rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID, err.Error())
		return nil, err
	}

	callBackReq := &httpReq.AlertCallBackRequest{
		AlertRuleUuid: rule.AlertRuleUUID,
		ChannelUuid:   rule.ChannelUUID,
		From:          oplog.SlowQueryAlertPlatform,
	}
	// 开始请求自身call_back接口
	go func() {
		if err = s.requestCallBackApi(callBackReq); err != nil {
			log.Errorf("alertSrv.DeleteRules requestCallBackApi error:%s", err.Error())
		}
	}()
	return rule, nil
}

func (s *Service) DisableAlertRuleAndChannel(rule *alert.RuleTab, channel *alert.ChannelTab, strategy *alert.StrategyTab) (err error) {
	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("alertSrv.DisableAlertRuleAndChannel", sysMetrics.GetStatus(err), time.Since(start))

	err = s.switchAlertRuleAndChannel(rule, channel, strategy, 1)
	return
}

func (s *Service) EnableAlertRuleAndChannel(rule *alert.RuleTab, channel *alert.ChannelTab, strategy *alert.StrategyTab) (err error) {
	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("alertSrv.EnableAlertRuleAndChannel", sysMetrics.GetStatus(err), time.Since(start))

	err = s.switchAlertRuleAndChannel(rule, channel, strategy, 0)
	return
}

func (s *Service) DeleteAlertRule(ruleId string) (err error) {
	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("alertSrv.DeleteAlertRule", sysMetrics.GetStatus(err), time.Since(start))

	err = s.monitorClient.DeleteAlertRule(ruleId)
	return
}

func (s *Service) DeleteAlertRulesAndChannelsAndStrategy(ruleIds, channelNames, strategyIds []string) (err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.DeleteAlertRulesAndChannelsAndStrategy", sysMetrics.GetStatus(err), time.Since(start))
	// strategy必须先与channel删除
	for _, strategyId := range strategyIds {
		if err = s.monitorClient.DeleteNotificationStrategy(strategyId); err != nil {
			log.Errorf("alertSrv.DeleteAlertRulesAndChannelsAndStrategy DeleteNotificationStrategy strategyId:%s ,error:%s", strategyId, err.Error())
			return err
		}
	}
	for _, ruleId := range ruleIds {
		if err = s.monitorClient.DeleteAlertRule(ruleId); err != nil {
			log.Errorf("alertSrv.DeleteAlertRulesAndChannelsAndStrategy DeleteAlertRule ruleId:%s ,error:%s", ruleId, err.Error())
			return err
		}
	}
	for _, channelName := range channelNames {
		if err = s.monitorClient.DeleteAlertChannel(channelName); err != nil {
			log.Errorf("alertSrv.DeleteAlertRulesAndChannelsAndStrategy DeleteAlertChannel channelName:%s ,error:%s", channelName, err.Error())
			return err
		}
	}
	return nil
}

func (s *Service) switchAlertRuleAndChannel(rule *alert.RuleTab, channel *alert.ChannelTab, strategy *alert.StrategyTab, isDisable int) (err error) {
	var (
		ruleVersion, chVersion, stVersion string
		remoteAlertRule                   *GetSingleAlertRuleResponse
		remoteChannel                     *GetSingleNotifyChannelResponse
		remoteStrategy                    *GetSingleNotificationStrategyResponse
	)
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.switchAlertRuleAndChannel", sysMetrics.GetStatus(err), time.Since(start))

	if remoteAlertRule, remoteChannel, remoteStrategy, err = s.getRemoteAlertRule(rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID); err != nil {
		log.Errorf("alertSrv.switchAlertRuleAndChannel getRemoteAlertRule monitorRuleID:%s,channelName:%s,strategyID:%s,error:%s", rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID, err.Error())
		return err
	}

	if remoteAlertRule.IsDelete || remoteChannel.IsDelete || remoteStrategy.IsDelete {
		s.deleteRemoteAlertRule(rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID)
		// 删除本地的rule，并记录下记录
		if err = alert.DeleteRuleAbout(s.db, rule, channel, strategy); err != nil {
			return err
		}
		if err = s.alertOpLog.Record(oplog.MonitorPlatform, rule.AlertRuleUUID, oplog.DELETE, oplog.DELETE_ALERT, rule.DBEnv, map[string]interface{}{
			"rule":     rule,
			"channel":  channel,
			"strategy": strategy,
		}, ""); err != nil {
			log.Warningf("alertSrv.switchAlertRuleAndChannel when update found rule has been deteled on monitor DeleteRule alertOpLog failed:%s", err.Error())
		}
		callBackReq := &httpReq.AlertCallBackRequest{
			AlertRuleUuid: rule.AlertRuleUUID,
			ChannelUuid:   rule.ChannelUUID,
			From:          oplog.SlowQueryAlertPlatform,
		}
		// 开始请求自身call_back接口
		go func() {
			if err = s.requestCallBackApi(callBackReq); err != nil {
				log.Errorf("alertSrv.switchAlertRuleAndChannel when delete requestCallBackApi error:%s", err.Error())
			}
		}()
		return fmt.Errorf("we detected an Alert Rule %s had been deleted on Monitoring Platform. Please re-create if you still need it", rule.AlertDisplayName)
	}

	if remoteAlertRule.Disable() && remoteChannel.Disable() && remoteStrategy.Disable() && isDisable == 1 {
		go s.execWebHook(rule.AlertRuleUUID, rule.ChannelUUID)
		return err
	}

	if ruleVersion, err = stringUtil.IncrementString(remoteAlertRule.DataVersion, DataVersionDelta); err != nil {
		return err
	}

	if chVersion, err = stringUtil.IncrementString(remoteChannel.DataVersion, DataVersionDelta); err != nil {
		return err
	}

	if stVersion, err = stringUtil.IncrementString(remoteStrategy.DataVersion, DataVersionDelta); err != nil {
		return err
	}

	// 剩下的全部是更新的了
	if _, err = s.monitorClient.UpdateAlertRule(rule.MonitorRuleID, &UpdateAlertRuleRequest{
		AlertRule: Rule{
			IsDisabled:  &isDisable,
			DataVersion: ruleVersion,
		},
		UpdateMask: UpdateMask{Paths: []string{"is_disabled"}},
	}); err != nil {
		if strings.Contains(err.Error(), "msg:Nothing updated!") == false {
			return err
		}
	}

	// 如果远端状态是disable， 就给置为enable,否则不需要更新
	if *remoteChannel.IsDisabled == 1 {
		isDisable = 0
		if _, err = s.monitorClient.UpdateChannel(channel.ChannelName, &UpdateNotifyChannelRequest{
			NotifyChannel: NotifyChannel{
				IsDisabled:  &isDisable,
				DataVersion: chVersion,
				Channel: Channel{
					Version: ChannelVersion,
				},
			},
			UpdateMask: UpdateMask{Paths: []string{"is_disabled"}},
		}); err != nil {
			if strings.Contains(err.Error(), "msg:Nothing updated!") == false {
				return err
			}
		}
	}
	// 如果远端状态是disable， 就给置为enable，否则不需要更新
	if *remoteStrategy.IsDisabled == 1 {
		isDisable = 0
		if _, err = s.monitorClient.UpdateStrategy(strategy.StrategyID, &UpdateNotificationStrategyRequest{
			AlertNotificationBind: NotificationBind{
				IsDisabled:  &isDisable,
				DataVersion: stVersion,
			},
			UpdateMask: UpdateMask{Paths: []string{"is_disabled"}},
		}); err != nil {
			if strings.Contains(err.Error(), "msg:Nothing updated!") == false {
				return err
			}
		}
	}
	return nil
}

func (s *Service) GetAlertRuleAndChannelDetail(alertRuleUUID string) (*alert.RuleTab, *alert.ChannelTab, error) {
	var (
		metaJson, usersJson []byte
		rule                *alert.RuleTab
		channel             *alert.ChannelTab
		strategy            *alert.StrategyTab
		remoteRule          *GetSingleAlertRuleResponse
		remoteChannel       *GetSingleNotifyChannelResponse
		remoteStrategy      *GetSingleNotificationStrategyResponse
		err                 error
	)
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.GetAlertRuleAndChannelDetail", sysMetrics.GetStatus(err), time.Since(start))

	// 1. fetch local rule,channel,strategy info
	if rule, channel, strategy, err = s.getLocalAlertRule(alertRuleUUID); err != nil {
		log.Errorf("")
		return nil, nil, err
	}

	// 2. fetch remote_monitor rule,channel,strategy info
	if remoteRule, remoteChannel, remoteStrategy, err = s.getRemoteAlertRule(rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID); err != nil {
		log.Errorf("alertSrv.GetAlertRuleAndChannelDetail getRemoteAlertRule monitorRuleID:%s,channelName:%s,strategyID:%s,error:%s", rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID, err.Error())
		return nil, nil, err
	}

	if err = s.syncStatusFromMonitor(rule, channel, strategy, remoteRule, remoteChannel, remoteStrategy); err != nil {
		log.Errorf("alertSrv.GetAlertRuleAndChannelDetail syncStatusFromMonitor monitorRuleID:%s,channelName:%s,strategyID:%s,error:%s", rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID, err.Error())
		return nil, nil, err
	}

	// update local data
	{
		// update local rule
		rule.Severity = remoteRule.AlertRuleContent.Severity
		rule.ForRange = remoteRule.AlertRuleContent.Content.For
		rule.EvaluationInterval = remoteRule.AlertRuleContent.Content.EvaluationInterval
		rule.AlarmMsg = remoteRule.AlertRuleContent.Content.MessageTemplate.FiringMessage
		rule.ResolveMsg = remoteRule.AlertRuleContent.Content.MessageTemplate.ResolvedMessage

		if len(remoteStrategy.ChannelConfig) > 0 && remoteStrategy.ChannelConfig[0].Config != nil {
			notification := &responseMeta.Notification{}
			if remoteStrategy.ChannelConfig[0].Config.Seatalk != nil && remoteStrategy.ChannelConfig[0].Config.Seatalk.Enabled {
				notification.Seatalk = &responseMeta.Seatalk{
					WebHook: []string{remoteChannel.Channel.Spec.ChannelInfo.Seatalk.SendTo[0].BotWebhookUrl},
				}
			}
			if remoteStrategy.ChannelConfig[0].Config.Email != nil && remoteStrategy.ChannelConfig[0].Config.Email.Enabled {
				notification.Email = &responseMeta.Email{}
			}
			if remoteStrategy.ChannelConfig[0].Config.Mattermost != nil && remoteStrategy.ChannelConfig[0].Config.Mattermost.Enabled {
				notification.MatterMost = &responseMeta.MatterMost{
					Channel: []string{remoteChannel.Channel.Spec.ChannelInfo.Mattermost.SendTo[0].MmChannel},
				}
			}
			if metaJson, err = json.Marshal(notification); err != nil {
				log.Warningf("json Marshal notification err:%s", err)
			} else {
				channel.MetaJson = string(metaJson)
			}
			if remoteStrategy.ChannelConfig[0].Config.Email.Enabled {
				if usersJson, err = json.Marshal(remoteChannel.Channel.Spec.ChannelInfo.Email.SendTo); err != nil {
					log.Warningf("json Marshal usersJson err:%s", err)
				} else {
					channel.UsersJson = string(usersJson)
				}
			}
		}

		if err = alert.UpdateRuleAndChannelAndStrategy(s.db, rule, channel, strategy); err != nil {
			log.Errorf("alertSrv.GetAlertRuleAndChannelDetail UpdateRuleAndChannelAndStrategy error:%s", err.Error())
			return nil, nil, err
		}
	}

	if err = s.alertOpLog.Record(oplog.SlowQueryAlertPlatform, rule.AlertRuleUUID, oplog.UPDATE, oplog.UPDATE_ALERT, rule.DBEnv, map[string]interface{}{
		"rule":     rule,
		"channel":  channel,
		"strategy": strategy,
	}, map[string]interface{}{
		"rule":     rule,
		"channel":  channel,
		"strategy": strategy,
	}); err != nil {
		log.Warningf("alertSrv.GetAlertRuleAndChannelDetail UpdateRuleAndChannelAndStrategy alterRuleUUID:%s,alertOpLog failed:%s", rule.AlertRuleUUID, err.Error())
	}

	return rule, channel, nil
}

// IsAdmin include dba and alert developer
func (s *Service) IsAdmin(c *gin.Context, user string) (bool, error) {
	var (
		err        error
		isDbaAdmin bool
	)
	for _, admin := range s.conf.Admins {
		if admin == user {
			return true, nil
		}
	}
	if isDbaAdmin, err = uic.IsDBaasAdmin(c, user); err != nil {
		return false, err
	}
	return isDbaAdmin, nil
}

func (s *Service) GetAlertRuleList(cond *httpReq.GetAlertRuleListRequest) ([]*alert.RuleAndChannel, error) {
	var (
		ruleAndChannels []*alert.RuleAndChannel
		remoteRule      *GetSingleAlertRuleResponse
		err             error
		wg              sync.WaitGroup
		start           = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("alertSrv.GetAlertRuleList", sysMetrics.GetStatus(err), time.Since(start))
	// 1. fetch local base info
	if ruleAndChannels, err = s.FindAlertByCond(cond); err != nil {
		return nil, err
	}

	// 2. fetch remote monitor base info
	for _, ruleAndChannel := range ruleAndChannels {
		wg.Add(1)
		go func(r *alert.RuleTab) {
			defer wg.Done()
			if remoteRule, err = s.monitorClient.GetSingleAlertRule(r.MonitorRuleID); err != nil {
				log.Errorf("alertSrv.GetAlertRuleList get remote data version failed, %v", err)
				return
			}

			// update local data
			{
				r.Severity = remoteRule.AlertRuleContent.Severity
				r.ForRange = remoteRule.AlertRuleContent.Content.For
				r.EvaluationInterval = remoteRule.AlertRuleContent.Content.EvaluationInterval
				r.AlarmMsg = remoteRule.AlertRuleContent.Content.MessageTemplate.FiringMessage
				r.ResolveMsg = remoteRule.AlertRuleContent.Content.MessageTemplate.ResolvedMessage

				if err = alert.UpdateRule(s.db, r); err != nil {
					log.Errorf("alertSrv.GetAlertRuleList UpdateRule error:%s", err.Error())
					return
				}
			}
		}(ruleAndChannel.RuleTab)
	}
	wg.Wait()
	return ruleAndChannels, nil
}

func (s *Service) GetMonitorEvent(ruleId, alertId, databaseName string, startTime, endTime int64) (resp *GetAlertEventResponse, err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.GetMonitorEvent", sysMetrics.GetStatus(err), time.Since(start))

	resp, err = s.monitorClient.GetAlertEvent(&GetAlertEventRequest{
		AuthProject:    ProjectName,
		AuthEventSetId: ProjectName,
		StartTime:      startTime,
		EndTime:        endTime,
		ProjectName:    ProjectName,
		RuleId:         ruleId,
		AlertId:        alertId,
		Annotations: map[string]string{
			"database_name": databaseName,
		},
	})
	return
}

func (s *Service) accordMutedUpdateStatus() (err error) {
	var messageList []*response.AlertMessage

	if messageList, err = s.ck.GetUnMutedAndTTLStatusMessageList(); err != nil {
		log.Errorf("GetUnMutedAndTTLStatusMessageList err:%s, time:%s", err, time.Now().String())
		return err
	}
	if err = s.daemonUpdateStatus(messageList); err != nil {
		log.Errorf("daemonUpdateStatus err:%s, time:%s", err, time.Now().String())
	}
	return err
}

func (s *Service) alwaysUpdateMuteDBStatus() (err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.alwaysUpdateMuteDBStatus", sysMetrics.GetStatus(err), time.Since(start))
	// 更新状态
	if err = s.ck.BatchUpdateMuteStatusToTTL(MuteMutedStatus, MuteTTLStatus, time.Now().Unix()); err != nil {
		log.Errorf("when status=muted and end_time < time.now.unix update status err:%s, time:%s", err, time.Now().String())
	}
	return err
}

func (s *Service) daemonUpdateStatus(messageList []*response.AlertMessage) (err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.daemonUpdateStatus", sysMetrics.GetStatus(err), time.Since(start))

	var eventData *GetAlertEventDetailResponse
	for _, message := range messageList {
		// 获取 monitor状态信息
		if eventData, err = s.monitorClient.GetAlertEventDetailByAlertId(message.MonitorAlertID, ProjectName, ProjectName); err != nil {
			log.Errorf("get event data error:%s", err.Error())
			continue
		}
		log.Infof("daemonUpdateStatus get monitor eventData:%v", eventData.Data)
		if eventData.Data.Status == storeReq.MonitorResolvedStatus || eventData.Data.Status == storeReq.MonitorClosedStatus {
			// 更新 状态
			if err = s.ck.UpdateAlertMessageStatus(message.MonitorAlertID, eventData.Data.Status); err != nil {
				log.Errorf("update alert status err:%s", err.Error())
				continue
			}
			// 删除mute
			if err = s.ck.DeleteMute(message.MonitorRuleID); err != nil {
				log.Errorf("delete alert mute err:%s", err.Error())
				continue
			}
		}
	}
	return nil
}

func (s *Service) requestCallBackApi(messages *httpReq.AlertCallBackRequest) error {
	var (
		err     error
		req     *http.Request
		payload []byte
	)
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.requestCallBackApi", sysMetrics.GetStatus(err), time.Since(start))

	if payload, err = json.Marshal(messages); err != nil {
		return err
	}

	if req, err = http.NewRequest(http.MethodPost, s.conf.CallBackUrl, bytes.NewBuffer(payload)); err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+s.conf.CallBackToken)
	if _, err = s.httpClient.Do(req); err != nil {
		log.Errorf("call back http err:%s", err.Error())
		return err
	}
	return nil
}

func (s *Service) getRemoteAlertRule(monitorRuleID, ChannelName, StrategyID string) (*GetSingleAlertRuleResponse, *GetSingleNotifyChannelResponse, *GetSingleNotificationStrategyResponse, error) {
	var (
		start           = time.Now()
		err             error
		remoteAlertRule *GetSingleAlertRuleResponse
		remoteChannel   *GetSingleNotifyChannelResponse
		remoteStrategy  *GetSingleNotificationStrategyResponse
	)

	defer sysMetrics.CollectStoreMetrics("alertSrv.getRemoteAlertRule", sysMetrics.GetStatus(err), time.Since(start))

	if remoteAlertRule, err = s.monitorClient.GetSingleAlertRule(monitorRuleID); err != nil {
		if errors.Is(err, MonitorNotFoundErr) {
			remoteAlertRule = &GetSingleAlertRuleResponse{Rule{IsDelete: true}}
		} else {
			log.Warningf("Get remote rule data version failed, %v", err)
			return remoteAlertRule, remoteChannel, remoteStrategy, err
		}
	}

	if remoteChannel, err = s.monitorClient.GetNotifyChannel(ChannelName); err != nil {
		if errors.Is(err, MonitorNotFoundErr) {
			remoteChannel = &GetSingleNotifyChannelResponse{NotifyChannel{IsDelete: true}}
		} else {
			log.Warningf("Get remote channel data version failed, %v", err)
			return remoteAlertRule, remoteChannel, remoteStrategy, err
		}
	}

	if remoteStrategy, err = s.monitorClient.GetSingleNotificationStrategy(StrategyID); err != nil {
		if errors.Is(err, MonitorNotFoundErr) {
			remoteStrategy = &GetSingleNotificationStrategyResponse{IsDelete: true}
		} else {
			log.Warningf("Get remote strategy data version failed, %v", err)
			return remoteAlertRule, remoteChannel, remoteStrategy, err
		}
	}
	return remoteAlertRule, remoteChannel, remoteStrategy, nil
}

func (s *Service) execWebHook(alertRuleUUID, channelUUID string) {
	var (
		err   error
		start = time.Now()
	)
	//拼凑消息体
	callBackReq := &httpReq.AlertCallBackRequest{
		AlertRuleUuid: alertRuleUUID,
		ChannelUuid:   channelUUID,
		From:          oplog.SlowQueryAlertPlatform,
	}
	// 开始请求自身call_back接口
	if err = s.requestCallBackApi(callBackReq); err != nil {
		log.Errorf("requestCallBackApi error:%s", err.Error())
	}
	sysMetrics.CollectStoreMetrics("alertSrv.execWebHook", sysMetrics.GetStatus(err), time.Since(start))
}

func (s *Service) deleteRemoteAlertRule(monitorRuleID, channelName, strategyID string) {
	var (
		err   error
		start = time.Now()
	)
	defer sysMetrics.CollectStoreMetrics("alertSrv.deleteRemoteAlertRule", sysMetrics.GetStatus(err), time.Since(start))

	if err = s.monitorClient.DeleteNotificationStrategy(strategyID); err != nil {
		if errors.Is(err, MonitorNotFoundErr) {
			log.Warningf("strategyID:%s has been deleted", strategyID)
		} else {
			log.Errorf("DeleteNotificationStrategy strategyID:%s,error:%s", strategyID, err.Error())
		}
	}

	if err = s.monitorClient.DeleteAlertRule(monitorRuleID); err != nil {
		if errors.Is(err, MonitorNotFoundErr) {
			log.Warningf("monitorRuleID:%s has been deleted", monitorRuleID)
		} else {
			log.Errorf("DeleteAlertRule strategyID:%s,error:%s", strategyID, err.Error())
		}
	}

	if err = s.monitorClient.DeleteAlertChannel(channelName); err != nil {
		if errors.Is(err, MonitorNotFoundErr) {
			log.Warningf("channelName:%s has been deleted", channelName)
		} else {
			log.Errorf("DeleteAlertChannel strategyID:%s,error:%s", strategyID, err.Error())
		}
	}
}

func (s *Service) getLocalAlertRule(alertRuleUUID string) (rule *alert.RuleTab, channel *alert.ChannelTab, strategy *alert.StrategyTab, err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("alertSrv.deleteRemoteAlertRule", sysMetrics.GetStatus(err), time.Since(start))

	if rule, err = s.FindAlertRuleFromDBByUUID(alertRuleUUID); err != nil {
		log.Errorf("service UpdateMonitorRule FindAlertRuleFromDBByUUID error:%s", err.Error())
		return
	}
	//if rule.ChannelUUID != channelUUID {
	//	return nil, nil, nil, fmt.Errorf("channel bing rule not found record, channel uuid=%s", channelUUID)
	//}

	if channel, err = s.FindAlertChannelFromDBByUUID(rule.ChannelUUID); err != nil {
		log.Errorf("service UpdateMonitorRule FindAlertChannelFromDBByUUID error:%s", err.Error())
		return
	}

	if strategy, err = s.FindStrategyFromDBByID(rule.StrategyID); err != nil {
		log.Errorf("service UpdateMonitorRule FindStrategyFromDBByID error:%s", err.Error())
		return
	}
	return
}

func (s *Service) syncStatusFromMonitor(rule *alert.RuleTab, channel *alert.ChannelTab, strategy *alert.StrategyTab, remoteRule *GetSingleAlertRuleResponse, remoteChannel *GetSingleNotifyChannelResponse, remoteStrategy *GetSingleNotificationStrategyResponse) (err error) {
	if remoteRule.IsDelete || remoteChannel.IsDelete || remoteStrategy.IsDelete {
		s.deleteRemoteAlertRule(rule.MonitorRuleID, channel.ChannelName, strategy.StrategyID)
		// 删除本地的rule，并记录下记录
		if err = alert.DeleteRuleAbout(s.db, rule, channel, strategy); err != nil {
			log.Errorf("delete rule from db,ruleUUID:%s,channelUUID:%s,strategyID:%s, error:%s", rule.AlertRuleUUID, channel.ChannelUUID, strategy.StrategyID, err.Error())
			return err
		}
		if lErr := s.alertOpLog.Record(oplog.MonitorPlatform, rule.AlertRuleUUID, oplog.DELETE, oplog.DELETE_ALERT, rule.DBEnv, map[string]interface{}{
			"rule":     rule,
			"channel":  channel,
			"strategy": strategy,
		}, ""); lErr != nil {
			log.Warningf("alertSrv.UpdateMonitorRule when update found rule has been deteled on monitor DeleteRule alertOpLog failed:%s", lErr.Error())
		}
		callBackReq := &httpReq.AlertCallBackRequest{
			AlertRuleUuid: rule.AlertRuleUUID,
			ChannelUuid:   rule.ChannelUUID,
			From:          oplog.SlowQueryAlertPlatform,
		}
		// 开始请求自身call_back接口
		go func() {
			if lErr := s.requestCallBackApi(callBackReq); lErr != nil {
				log.Errorf("alertSrv.UpdateMonitorRule when delete requestCallBackApi error:%s", lErr.Error())
			}
		}()
		return fmt.Errorf("alertSrv.UpdateMonitorRule we detected an Alert Rule %s had been deleted on Monitoring Platform. Please re-create if you still need it", rule.AlertDisplayName)
	}

	//update操作时检测到远端rule/channel/strategy均为disable状态，说明用户已在monitor平台将报警规则disable调,此处需要模拟callback补偿alertmessage的状态更新
	if remoteRule.Status() != rule.Status {
		switch rule.Status {
		case httpReq.DISABLE:
			// do nothing
		case httpReq.ENABLE:
			if remoteChannel.Disable() && remoteStrategy.Disable() {
				go s.execWebHook(rule.AlertRuleUUID, rule.ChannelUUID)
				log.Infof("alertSrv.UpdateMonitorRule remote rule has disabled, send webhook for alert message's status consistent")
			}
		}
		rule.Status = remoteRule.Status()
	}
	return
}

func (s *Service) AlertMessageCallback(callbackReq *httpReq.AlertCallBackRequest) error {
	var (
		err, remoteErr, escapeErr error
		labelInfoBytes            []byte
		message                   *response.AlertMessage
		messages                  []*response.AlertMessage
		template                  *conf.AlertTemplate
		remoteRule                *GetSingleAlertRuleResponse
		ruleInfo                  *alert.RuleTab
		channelInfo               *alert.ChannelTab
		start                     = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("alertSrv.SaveAlertMessage", sysMetrics.GetStatus(err), time.Since(start))

	/**************************开始处理报警消息**************************/
	for _, alertInfo := range callbackReq.AlertList {
		// 将encode的链接编码转回
		splitMessage := strings.Split(alertInfo.Message, splitMessageKey)
		if len(splitMessage) == 2 {
			splitMessage[1], escapeErr = url.QueryUnescape(splitMessage[1])
			if escapeErr != nil {
				log.Warningf("cant`t from %s find url encode,just find %s", alertInfo.Message, splitMessage[1])
			}
			alertInfo.Message = strings.Join(splitMessage, splitMessageKey)
		}
		// 先去查询数据库中是否有事件id
		if message, err = s.ck.GetAlertMessageByAlertID(alertInfo.AlertId); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("alertSrv.SaveAlertMessage GetAlertMessageByAlertID error:%s", err.Error())
			return err
		}

		// 第一次报警
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if labelInfoBytes, err = json.Marshal(alertInfo.Labels); err != nil {
				log.Errorf("alertSrv.SaveAlertMessage Marshal labelInfoBytes error:%s", err.Error())
				return err
			}

			if template, err = s.FindAlertTemplateByUUID(callbackReq.AlertTemplateUUID); err != nil {
				log.Errorf("alertSrv.SaveAlertMessage FindAlertTemplateByUUID error:%s", err.Error())
				return err
			}

			alertMessage := &storeReq.AlertMessage{
				MonitorAlertID: fmt.Sprintf("%d", alertInfo.AlertId),
				MonitorRuleID:  alertInfo.AlertRuleId,
				AlertRuleUUID:  callbackReq.AlertRuleUuid,
				AlertRuleName:  callbackReq.AlertName,
				AlertStrategy:  callbackReq.AlertStrategy,
				ChannelUUID:    callbackReq.ChannelUuid,
				CMDB:           alertInfo.Labels.CMDB,
				DataBaseName:   alertInfo.Labels.DatabaseName,
				Env:            alertInfo.Labels.DbEnv,
				Status:         getAlertStatus(alertInfo.Status),
				Severity:       alertInfo.Severity,
				Message:        alertInfo.Message,
				ACKBy:          alertInfo.AckInfo,
				LabelInfo:      string(labelInfoBytes),
				TemplateName:   template.Name,
				AlertCount:     1,
				StartTime:      alertInfo.StartTime, //使用monitor平台webhook的startTime时间戳，解决不同系统间时钟差异对，查询报警消息状态接口调用结果的影响
				LastAlertTime:  time.Now(),
				ResolveTime:    0,
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			}

			if err = s.ck.CreateAlertMessage(alertMessage); err != nil {
				log.Errorf("alertSrv.SaveAlertMessage CreateAlertMessage error:%s", err.Error())
			}
		} else {
			// 第n次报警,本地状态非closed即可更新
			if !(message.Status == storeReq.MonitorClosedStatus || message.Status == storeReq.AlertResolved) {
				if err = s.ck.UpdateAlertMessage(&storeReq.AlertMessage{
					MonitorAlertID: message.MonitorAlertID,
					MonitorRuleID:  message.MonitorRuleID,
					AlertStrategy:  callbackReq.AlertStrategy,
					AlertRuleUUID:  message.AlertRuleUUID,
					AlertRuleName:  message.AlertRuleName,
					ChannelUUID:    message.ChannelUUID,
					CMDB:           message.CMDB,
					DataBaseName:   message.DataBaseName,
					Env:            message.Env,
					Status:         getAlertStatus(alertInfo.Status),
					Severity:       message.Severity,
					Message:        message.Message,
					ACKBy:          message.ACKBy,
					LabelInfo:      message.LabelInfo,
					TemplateName:   message.TemplateName,
					AlertCount:     message.AlertCount + 1,
					StartTime:      message.StartTime,
					ResolveTime:    getResolveTime(alertInfo.Status),
					LastAlertTime:  time.Now(),
					CreateTime:     message.CreateTime,
					UpdateTime:     time.Now(),
				}); err != nil {
					log.Errorf("alertSrv.SaveAlertMessage UpdateAlertMessage error:%s", err.Error())
				}
			} else {
				log.Warningf("alertSrv.SaveAlertMessage received callback alert.status:%s,message.stats:%s,alertId:%s,ruleId:%s", alertInfo.Status, message.Status, message.MonitorAlertID, message.MonitorRuleID)
			}
		}
		//pending消息兜底逻辑：解决个别callback回调消息丢失
		if alertInfo.Status == storeReq.AlertResolved || alertInfo.Status == storeReq.AlertClosed {
			if messages, err = s.ck.GetAlertMessageByMonitorRuleIdAndStatus(callbackReq.AlertRuleUuid, storeReq.AlertPending); err != nil {
				log.Warningf("alertSrv.SaveAlertMessage GetAlertMessageByMonitorRuleIdAndStatus error:%s, alert_id:%s, status:%s", err.Error(), ruleInfo.MonitorRuleID, storeReq.AlertPending)
				continue
			}

			for _, m := range messages {
				if alertInfo.Status == storeReq.MonitorResolvedStatus && m.DataBaseName != alertInfo.Labels.DatabaseName {
					continue
				}
				alterID, err := strconv.ParseInt(m.MonitorAlertID, 10, 64)
				if err != nil {
					continue
				}
				if alertInfo.AlertId > alterID {
					if err = s.ck.UpdateAlertMessageStatus(m.MonitorAlertID, alertInfo.Status); err != nil {
						log.Warningf("alertSrv.SaveAlertMessage UpdateAlertMessageStatus error:%s", err.Error())
					}
					log.Infof("alertSrv.SaveAlertMessage %s become %s to %s", m.MonitorAlertID, m.Status, storeReq.AlertClosed)
				}
			}
		}
	}
	/**************************处理报警消息结束**************************/
	/*************************开始处理报警规则**************************/
	// 获取本地状态
	if ruleInfo, err = s.FindAlertRuleFromDBByUUID(callbackReq.AlertRuleUuid); err != nil {
		// 这里，可能会出现 用户在更新时候发现远端rule不在了=>那个时候规则被删除了 的情况， 出现这种情况是monitor端删除了报警的情况，
		// 因为如果是我们自己平台删除的，需要我们给用户提供列表中有他,但是并不会，我们做了过滤
		// 所以，这种情况只可能是远端被删除了， 我们通过更新等接口发现后,给他后置删除的，所以我们不需要处理，直接返回即可
		log.Errorf("alertSrv.SaveAlertMessage back message`s judge alertRule alertRuleUuid:%s ,error:%s", callbackReq.AlertRuleUuid, err.Error())
		return err
	}

	if channelInfo, err = s.FindAlertChannelFromDBByUUID(ruleInfo.ChannelUUID); err != nil {
		log.Errorf("alertSrv.SaveAlertMessage FindAlertChannelFromDBByUUID channelUUID:%s,  error:%s", ruleInfo.ChannelUUID, err.Error())
		return err
	}

	if remoteRule, remoteErr = s.monitorClient.GetSingleAlertRule(ruleInfo.MonitorRuleID); remoteErr != nil && !errors.Is(remoteErr, MonitorNotFoundErr) {
		log.Errorf("alertSrv.SaveAlertMessage GetSingleAlertRule monitorRuleID:%s ,error:%s", ruleInfo.MonitorRuleID, remoteErr.Error())
		return remoteErr
	}
	// 检测&维护 slow query与monitor两个平台间报警规则一致性
	if callbackReq.FromMonitor() && !callbackReq.HasMessageClosed() && ruleInfo.Status == httpReq.DISABLE && !remoteRule.Disable() {
		// 更新本地状态
		if err = alert.UpdateRuleStatus(s.db, &alert.RuleTab{
			AlertRuleUUID: ruleInfo.AlertRuleUUID,
			Status:        httpReq.ENABLE,
			Modifier:      oplog.MonitorPlatform,
		}); err != nil {
			log.Errorf("alertSrv.SaveAlertMessage monitorClient UpdateRuleStatus alertRuleUUID:%s, status:%s, error:%s", ruleInfo.AlertRuleUUID, httpReq.ENABLE, err.Error())
			return err
		}
		// 查一下本地记录，我们应该是未删除的
		if err = s.alertOpLog.Record(oplog.MonitorPlatform, message.AlertRuleUUID, oplog.CREATE_ALERT, oplog.UPDATE_ALERT_STATUS, message.Env, httpReq.DISABLE, httpReq.ENABLE); err != nil {
			log.Warningf("alertSrv.SaveAlertMessage alertOpLog monitorRuleID:%s ,actionType:%s ,actionName:%s ,oldStatus:%s ,newStatus:%s ,failed error:%s", message.MonitorAlertID, oplog.CREATE_ALERT, oplog.UPDATE_ALERT_STATUS, httpReq.DISABLE, httpReq.ENABLE, err.Error())
		}
	}

	if callbackReq.HasMessageClosed() {
		//主动删除报警规则
		if errors.Is(remoteErr, MonitorNotFoundErr) {
			// 如果是查询到的结果是404 not found,说明远端被删除了，我们需要删除下本地记录下日志
			if remoteErr = s.monitorClient.DeleteNotificationStrategy(ruleInfo.StrategyID); remoteErr != nil {
				log.Errorf("alertSrv.SaveAlertMessage DeleteNotificationStrategy StrategyID:%s ,error:%s", ruleInfo.StrategyID, remoteErr.Error())
			}
			if remoteErr = s.monitorClient.DeleteAlertChannel(channelInfo.ChannelName); remoteErr != nil {
				log.Errorf("alertSrv.SaveAlertMessage DeleteAlertChannel ChannelName:%s ,error:%s", channelInfo.ChannelName, remoteErr.Error())
			}
			if callbackReq.FromMonitor() {
				// 删除本地rule/channel/strategy
				if err = alert.DeleteRuleAbout(s.db, ruleInfo, &alert.ChannelTab{ChannelUUID: ruleInfo.ChannelUUID}, &alert.StrategyTab{StrategyID: ruleInfo.StrategyID}); err != nil {
					log.Errorf("alertSrv.SaveAlertMessage DeleteRuleAbout ruleUUID:%s ,error:%s", ruleInfo.AlertRuleUUID, err.Error())
					return err
				}
			}

			if err = s.alertOpLog.Record(callbackReq.From, message.AlertRuleUUID, oplog.DELETE, oplog.DELETE_ALERT, message.Env, ruleInfo, ""); err != nil {
				log.Warningf("alertSrv.SaveAlertMessage alertOpLog failed error:%s", err.Error())
			}
		}

		// 2. 报警规则状态一致性校验： 判断下远端状态，如果与本地数据库记录不一致，以远端的为准去更新
		if ruleInfo.Status != remoteRule.Status() {
			if err = alert.UpdateRuleStatus(s.db, &alert.RuleTab{
				AlertRuleUUID: ruleInfo.AlertRuleUUID,
				Status:        remoteRule.Status(),
				Modifier:      callbackReq.From,
			}); err != nil {
				log.Errorf("alertSrv.SaveAlertMessage UpdateRuleStatus error:%s", err.Error())
				return err
			}
			if err = s.alertOpLog.Record(callbackReq.From, message.AlertRuleUUID, oplog.UPDATE, oplog.UPDATE_ALERT_STATUS, message.Env, map[string]string{"oldStatus": ruleInfo.Status}, map[string]string{"newStatus": remoteRule.Status()}); err != nil {
				log.Warningf("alertSrv.SaveAlertMessage callback change alert rule status alertOpLog failed error:%s", err.Error())
			}
		}
	}
	/*************************处理报警规则结束**************************/
	return nil
}

func getAlertStatus(status string) string {
	if strings.ToLower(status) == storeReq.AlertResolved {
		return storeReq.AlertResolved
	}
	if strings.ToLower(status) == storeReq.AlertClosed {
		return storeReq.AlertClosed
	}
	return storeReq.AlertPending
}

func getResolveTime(status string) uint64 {
	if strings.ToLower(status) == storeReq.AlertResolved {
		return uint64(time.Now().Unix())
	}
	return 0
}
