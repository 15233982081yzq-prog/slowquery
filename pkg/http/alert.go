package http

import (
	conf "smart-slowquery/conf/alert"
	alertMeta "smart-slowquery/internal/model/alert"
	cmdbUtil "smart-slowquery/internal/util/cmdb"
	requestMeta "smart-slowquery/pkg/store/request"
	responseMeta "smart-slowquery/pkg/store/response"

	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/oplog"
	"smart-slowquery/pkg/service/alert"

	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MuteForever        = "876000h"
	AlertMaxDBCount    = 100
	BatchDealCount     = 10
	maxSqlFilterLength = 200 * 1024
)

func (api *Api) CreateAlertRule(c *gin.Context) {
	var (
		err error
		//dbs          []string
		ruleManagers []*alert.RuleManager
		alertRuleVo  = &response.AlertVo{}
		req          = &request.AlertCreateRequest{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if len(req.DBS) > AlertMaxDBCount {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("create alert add db max count is %d", AlertMaxDBCount))
		return
	}

	req.Applicant = api.GetOperatorEmail(c)
	req.Status = request.ENABLE // 默认是enable

	//if dbs, err = api.dataBaseSrv.GetDataBases(req.CMDB, api.GetToken(c), strings.Split(req.DBEnv, ",")); err != nil {
	//	log.Errorf("http GetLogicDBByService cmdbUtil.ListLogicalDBService failed,cmdb_service:%s,db_env:%s ,error:%s", req.CMDB, req.DBEnv, err.Error())
	//	response.ToResponse(c, map[string]string{}, err)
	//	return
	//}
	//
	//if exist, diffs := stringUtil.IsSubset(req.DBS, dbs); !exist {
	//	response.ToResponse(c, map[string]string{}, fmt.Errorf("dbs:%v not in cmdb:%s, please check it", diffs, req.CMDB))
	//	return
	//}

	req.DBGroup = [][]string{req.DBS}
	if ruleManagers, err = api.AlertSrv.GenMgrRuleRequest(req.AlertCommonRequest); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if alertRuleVo.AlertUUIDs, alertRuleVo.ChannelUUIDs, err = api.AlertSrv.CreateMonitorRule(ruleManagers); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, alertRuleVo, nil)
}

func (api *Api) UpdateAlertRule(c *gin.Context) {

	var (
		ruleUUID, channelUUID string
		err                   error
		//dbs                   []string
		ruleManagers []*alert.RuleManager
		req          = &request.AlertUpdateRequest{}
	)
	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if len(req.DBS) > AlertMaxDBCount {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("update alert add db max count is %d", AlertMaxDBCount))
		return
	}

	req.Applicant = api.GetOperatorEmail(c)

	//if dbs, err = api.dataBaseSrv.GetDataBases(req.CMDB, api.GetToken(c), strings.Split(req.DBEnv, ",")); err != nil {
	//	log.Errorf("http GetLogicDBByService cmdbUtil.ListLogicalDBService failed,cmdb_service:%s,db_env:%s ,error:%s", req.CMDB, req.DBEnv, err.Error())
	//	return
	//}
	//
	//if exist, diffs := stringUtil.IsSubset(req.DBS, dbs); !exist {
	//	response.ToResponse(c, map[string]string{}, fmt.Errorf("dbs:%v not in cmdb:%s, please check it", diffs, req.CMDB))
	//	return
	//}

	if req.Status != "" && !(req.Status == request.ENABLE || req.Status == request.DISABLE) {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("req.status is err, status:%s", req.Status))
		return
	}
	req.DBGroup = [][]string{req.DBS}

	if ruleManagers, err = api.AlertSrv.GenMgrRuleRequest(req.AlertCommonRequest); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if ruleUUID, channelUUID, err = api.AlertSrv.UpdateMonitorRule(ruleManagers[0], req.RuleUUID); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, &response.UpdateAlertVo{AlertUUID: ruleUUID, ChannelUUID: channelUUID}, nil)
}

func (api *Api) UpdateAlertRuleStatus(c *gin.Context) {

	var (
		err       error
		oldStatus string
		req       = &request.ChangeAlertRuleStatusRequest{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	req.Applicant = api.GetOperatorEmail(c)

	if req.Status != "" && !(req.Status == request.ENABLE || req.Status == request.DISABLE) {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("req.status is err, status:%s", req.Status))
		return
	}

	if oldStatus, _, err = api.AlertSrv.ChangeRuleStatus(req.RuleUuid, req.Status, req.Applicant); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	// record operator log
	if err = api.alertOpLog.Record(req.Applicant, req.RuleUuid, oplog.UPDATE, oplog.UPDATE_ALERT_STATUS, req.DBEnv, map[string]string{"oldStatus": oldStatus}, map[string]string{"newStatus": req.Status}); err != nil {
		log.Warningf("UpdateAlertRuleStatus alertOpLog error:%s", err.Error())
	}

	response.ToResponse(c, &response.UpdateAlertVo{AlertUUID: req.RuleUuid}, nil)
}

func (api *Api) BatchUpdateAlertRulesStatus(c *gin.Context) {

	var (
		err                 error
		wg                  sync.WaitGroup
		mux                 sync.Mutex
		before, after       = make(map[string]string), make(map[string]string)
		batchUpdateStatusVo = &response.BatchUpdateStatusVo{}
		req                 = &request.BatchChangeAlertRuleStatusRequest{}
		actionType          = oplog.UPDATE_ALERT_STATUS
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	req.Applicant = api.GetOperatorEmail(c)
	if len(req.RuleUuids) > BatchDealCount {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("batch update count is %d", BatchDealCount))
		return
	}

	if req.Status != "" && !(req.Status == request.ENABLE || req.Status == request.DISABLE) {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("req.status is err, status:%s", req.Status))
		return
	}

	for _, alertUuid := range req.RuleUuids {
		wg.Add(1)
		go func(uuid string) {
			oldStatus, alertName, e := api.AlertSrv.ChangeRuleStatus(uuid, req.Status, req.Applicant)
			mux.Lock()
			if e != nil {
				log.Errorf("UpdateAlertRulesStatus AlertSrv.ChangeRuleStatus uuid:%s,status:%s,applicant:%s,error:%s", uuid, req.Status, req.Applicant, err.Error())
				batchUpdateStatusVo.FailRules = append(batchUpdateStatusVo.FailRules, response.FailInfo{
					UUID:      uuid,
					ErrMsg:    e.Error(),
					AlertName: alertName,
				})
			} else {
				batchUpdateStatusVo.SuccessRules = append(batchUpdateStatusVo.SuccessRules, uuid)
				before[uuid] = oldStatus
				after[uuid] = req.Status
			}
			mux.Unlock()
			wg.Done()
		}(alertUuid)
	}
	wg.Wait()

	if len(batchUpdateStatusVo.SuccessRules) > 1 {
		actionType = oplog.BATCH_UPDATE_ALERT_STATUS
	}

	if len(batchUpdateStatusVo.SuccessRules) > 0 {
		if err = api.alertOpLog.Record(req.Applicant, strings.Join(batchUpdateStatusVo.SuccessRules, ","),
			oplog.UPDATE, actionType,
			req.DBEnv, before, after); err != nil {
			log.Warningf("UpdateAlertRulesStatus alertOpLog error:%s", err.Error())
		}
	}

	if len(batchUpdateStatusVo.FailRules) > 0 {
		response.ToResponse(c, batchUpdateStatusVo, fmt.Errorf("batch update status to %s error, alertNames:[%s], please retry", req.Status, strings.Join(batchUpdateStatusVo.FailAlertNames(), ",")))
		return
	}

	response.ToResponse(c, batchUpdateStatusVo, nil)
}

func (api *Api) DeleteRules(c *gin.Context) {
	var (
		err                 error
		wg                  sync.WaitGroup
		mux                 sync.Mutex
		oldRuleRecord       = make(map[string]interface{})
		batchDeleteStatusVo = &response.BatchUpdateStatusVo{}
		req                 = &request.DeleteRulesRequest{}
		actionType          = oplog.DELETE_ALERT
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	req.Applicant = api.GetOperatorEmail(c)

	if len(req.RuleUuids) > BatchDealCount {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("batch delete count is %d", BatchDealCount))
		return
	}

	for _, alertUuid := range req.RuleUuids {
		wg.Add(1)
		go func(uuid string) {
			oldRule, e := api.AlertSrv.DeleteRules(uuid)

			mux.Lock()
			if e != nil {
				log.Errorf("DeleteRules AlertSrv.DeleteRules uuid:%s,error:%s", uuid, err.Error())
				batchDeleteStatusVo.FailRules = append(batchDeleteStatusVo.FailRules, response.FailInfo{
					UUID:      uuid,
					ErrMsg:    e.Error(),
					AlertName: oldRule.AlertDisplayName,
				})
			} else {
				batchDeleteStatusVo.SuccessRules = append(batchDeleteStatusVo.SuccessRules, uuid)
				oldRuleRecord[uuid] = oldRule
			}
			mux.Unlock()
			wg.Done()
		}(alertUuid)
	}
	wg.Wait()

	if len(batchDeleteStatusVo.SuccessRules) > 1 {
		actionType = oplog.BATCH_DELETE_ALERT
	}

	if len(batchDeleteStatusVo.SuccessRules) > 0 {
		if err = api.alertOpLog.Record(req.Applicant, strings.Join(batchDeleteStatusVo.SuccessRules, ","), oplog.DELETE, actionType, req.DBEnv, oldRuleRecord, ""); err != nil {
			log.Warningf("DeleteRules alertOpLog error:%s", err.Error())
		}
	}

	if len(batchDeleteStatusVo.FailRules) > 0 {
		response.ToResponse(c, batchDeleteStatusVo, fmt.Errorf("batch delete rule error alertNames:[%s], please retry", strings.Join(batchDeleteStatusVo.FailAlertNames(), ",")))
		return
	}

	response.ToResponse(c, batchDeleteStatusVo, nil)
}

func (api *Api) GetAlertRule(c *gin.Context) {
	var (
		err                error
		ruleCount          int64
		isAdmin            bool
		cmdbServices       []string
		ruleAndChannels    []*alertMeta.RuleAndChannel
		req                = &request.GetAlertRuleListRequest{}
		getAlertRuleListVo = &response.GetAlertRuleListResponse{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	// 用户没有选择任何cmdb
	if req.CMDBS == nil || len(req.CMDBS) == 0 {
		cmdbServices, err = cmdbUtil.GetServiceTree(c, c.Request.Header.Get("Authorization"), "shopee.", conf.GlobalConfig.SpaceConfig.SpaceHost)
		if err != nil {
			response.ToResponse(c, map[string]string{}, err)
			return
		} else {
			if len(strings.Join(cmdbServices, ",")) > maxSqlFilterLength {
				log.Infof("fetch user cmdb too long:%v", cmdbServices)
				// 超过一定长度 会导致sql过长无法查询，特殊处理
				req.CMDBS = []string{}
			} else {
				req.CMDBS = cmdbServices
			}
		}

		if isAdmin, err = api.AlertSrv.IsAdmin(c, req.Applicant); err != nil {
			response.ToResponse(c, map[string]string{}, err)
			return
		}

		if isAdmin {
			req.CMDBS = []string{}
		}
	}

	if req.Status != "" && !(req.Status == request.ENABLE || req.Status == request.DISABLE) {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("req.status is err, status:%s", req.Status))
		return
	}

	req.Applicant = api.GetOperatorEmail(c)

	if req.EndTime < req.StartTime {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("end time do not less than start time"))
		return
	}

	if ruleCount, err = api.AlertSrv.FindAlertCountByCond(req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	getAlertRuleListVo.TotalNum = int(ruleCount)
	getAlertRuleListVo.TotalPage = int(ruleCount)/req.PageSize + 1
	getAlertRuleListVo.PageSize = req.PageSize

	if int64((req.Page-1)*req.PageSize) > ruleCount-1 {
		response.ToResponse(c, getAlertRuleListVo, nil)
		return
	}

	if ruleAndChannels, err = api.AlertSrv.GetAlertRuleList(req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	getAlertRuleListVo.Rules = make([]*response.AlertInfoVo, 0)
	for _, ruleAndChannel := range ruleAndChannels {
		var (
			dbs, users []string
		)
		if e := json.Unmarshal([]byte(ruleAndChannel.DbsJson), &dbs); e != nil {
			log.Errorf("GetAlertRuleList Unmarshal DbsJson error:%s", err.Error())
			continue
		}
		if e := json.Unmarshal([]byte(ruleAndChannel.UsersJson), &users); e != nil {
			log.Errorf("GetAlertRuleList Unmarshal UserJson error:%s", err.Error())
			continue
		}
		getAlertRuleListVo.Rules = append(getAlertRuleListVo.Rules, &response.AlertInfoVo{
			CMDB:            ruleAndChannel.CMDB,
			DBEnv:           ruleAndChannel.DBEnv,
			DBS:             dbs,
			Status:          ruleAndChannel.RuleTab.Status,
			AlertName:       ruleAndChannel.AlertDisplayName,
			AlertUUID:       ruleAndChannel.AlertRuleUUID,
			Creator:         ruleAndChannel.Creator,
			EvaluateEvery:   ruleAndChannel.EvaluationInterval,
			Expression:      ruleAndChannel.Expression,
			ForRange:        ruleAndChannel.ForRange,
			ExpressionValue: ruleAndChannel.ExpressionValue,
			Severity:        ruleAndChannel.Severity,
			ChannelType:     ruleAndChannel.ChannelType,
			CreateTime:      ruleAndChannel.RuleTab.CreateTime,
			TemplateName:    ruleAndChannel.TemplateName,
			UpdateTime:      ruleAndChannel.RuleTab.UpdateTime,
			Recipient:       users,
		})
	}

	response.ToResponse(c, getAlertRuleListVo, nil)
}

func (api *Api) GetAlertRuleDetail(c *gin.Context) {
	var (
		err                              error
		alertTemplateUUID, alertRuleUuid string
		dbs, userList                    []string
		rule                             *alertMeta.RuleTab
		channel                          *alertMeta.ChannelTab
		alertDetailVo                    *response.AlertDetailVo
		channelConfig                    = &response.Notification{}
	)

	if alertRuleUuid = c.Query("rule_uuid"); alertRuleUuid == "" {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("alert rule uuid not found, uuid=%s", alertRuleUuid))
		return
	}

	if rule, channel, err = api.AlertSrv.GetAlertRuleAndChannelDetail(alertRuleUuid); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if err = json.Unmarshal([]byte(rule.DbsJson), &dbs); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if err = json.Unmarshal([]byte(channel.UsersJson), &userList); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	for _, templateInfo := range api.AlertSrv.GetAlertTemplates() {
		if templateInfo.Name == rule.TemplateName {
			alertTemplateUUID = templateInfo.UUID
			break
		}
	}

	if err = json.Unmarshal([]byte(channel.MetaJson), channelConfig); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	alertDetailVo = &response.AlertDetailVo{
		RuleUUID:    rule.AlertRuleUUID,
		ChannelUUID: rule.ChannelUUID,
		CMDB:        rule.CMDB,
		DBEnv:       rule.DBEnv,
		DBS:         dbs,
		Status:      rule.Status,
		Channel: &response.ChannelInfo{
			Recipient: response.Recipient{
				Dod:      channel.DodID,
				Users:    userList,
				Interval: channel.Interval,
			},
			Notification: *channelConfig,
		},
		Rule: &response.AlertRuleInfo{
			Name:              rule.AlertDisplayName,
			Trigger:           rule.Trigger,
			AlertTemplateUUID: alertTemplateUUID,
			Expression:        rule.Expression,
			ExpressionValue:   rule.ExpressionValue,
			ForRange:          rule.ForRange,
			EvaluateEvery:     rule.EvaluationInterval,
			Severity:          rule.Severity,
			AlertMsg:          rule.AlarmMsg,
			ResolveMsg:        rule.ResolveMsg,
		},
	}
	response.ToResponse(c, alertDetailVo, nil)
}

//----------------------------------------------------------//

func (api *Api) CreateAlertRuleTemplate(c *gin.Context) {
	var getTemplatesVo = &response.GetTemplatesVo{}

	getTemplatesVo.Templates = api.AlertSrv.GetAlertTemplates()

	response.ToResponse(c, getTemplatesVo, nil)
}

func (api *Api) GetAlertMessage(c *gin.Context) {
	var (
		err                   error
		messageCount          int
		isAdmin               bool
		cmdbServices          []string
		messageList           []*responseMeta.AlertMessageWithMuteOperator
		req                   = &request.GetAlertMessageListRequest{}
		getAlertMessageListVo = &response.GetAlertMessageListResponse{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if req.CMDBs == nil || len(req.CMDBs) == 0 {
		cmdbServices, err = cmdbUtil.GetServiceTree(c, c.Request.Header.Get("Authorization"), "shopee.", conf.GlobalConfig.SpaceConfig.SpaceHost)
		if err != nil {
			response.ToResponse(c, map[string]string{}, err)
			return
		} else {
			if len(strings.Join(cmdbServices, ",")) > maxSqlFilterLength {
				log.Infof("fetch user cmdb too long:%v", cmdbServices)
				// 超过一定长度 会导致sql过长无法查询，特殊处理
				req.CMDBs = []string{}
			} else {
				req.CMDBs = cmdbServices
			}
		}
	}

	if req.Status != "" && !(req.Status == requestMeta.AlertPending || req.Status == requestMeta.AlertResolved || req.Status == requestMeta.AlertClosed) {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("req.status is err, status:%s", req.Status))
		return
	}

	req.Applicant = api.GetOperatorEmail(c)

	if isAdmin, err = api.AlertSrv.IsAdmin(c, req.Applicant); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if isAdmin {
		req.CMDBs = []string{}
	}

	if messageCount, err = api.AlertSrv.FindAlertMessageListCountByCond(req); err != nil {
		log.Errorf("GetAlertMessage FindAlertMessageCountByCond error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	getAlertMessageListVo.TotalNum = messageCount
	getAlertMessageListVo.TotalPage = messageCount/req.PageSize + 1
	getAlertMessageListVo.PageSize = req.PageSize

	if messageList, err = api.AlertSrv.FindAlertMessageByCond(req); err != nil {
		log.Errorf("GetAlertMessage FindAlertMessageByCond error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	for _, message := range messageList {
		alertVo := &response.AlertMessageInfoVo{
			AlertID:       message.MonitorAlertID,
			AlertRuleName: message.AlertRuleName,
			TemplateName:  message.TemplateName,
			DBName:        message.DataBaseName,
			CMDB:          message.CMDB,
			Severity:      message.Severity,
			FirsTime:      message.CreateTime,
			LastTime:      message.LastAlertTime,
			Duration:      message.LastAlertTime.Sub(message.CreateTime).String(),
			Status:        message.Status,
			Message:       message.Message,
			Count:         int(message.AlertCount),
			AlertStrategy: message.AlertStrategy,
			MuteOperator:  message.MuteOperator,
			MuteTo:        message.MuteTo,
		}
		if message.Status == requestMeta.AlertPending {
			now := time.Now()
			alertVo.Duration = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Local().Location()).Sub(message.CreateTime).String()
		}
		getAlertMessageListVo.Alerts = append(getAlertMessageListVo.Alerts, alertVo)
	}

	response.ToResponse(c, getAlertMessageListVo, nil)
}

func (api *Api) GetAlertMessageAbstract(c *gin.Context) {
	var (
		err                       error
		isAdmin                   bool
		cmdbServices              []string
		req                       = &request.GetAlertMessageAbstractRequest{}
		getAlertMessageAbstractVo = &response.GetAlertMessageAbstractResponse{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if req.CMDBs == nil || len(req.CMDBs) == 0 {
		cmdbServices, err = cmdbUtil.GetServiceTree(c, c.Request.Header.Get("Authorization"), "shopee.", conf.GlobalConfig.SpaceConfig.SpaceHost)
		if err != nil {
			response.ToResponse(c, map[string]string{}, err)
			return
		} else {
			if len(strings.Join(cmdbServices, ",")) > maxSqlFilterLength {
				// 超过一定长度 会导致sql过长无法查询，特殊处理
				log.Infof("fetch user cmdb too long:%v", cmdbServices)
				req.CMDBs = []string{}
			} else {
				req.CMDBs = cmdbServices
			}
		}
	}

	req.Applicant = api.GetOperatorEmail(c)

	start := time.Now()
	if isAdmin, err = api.AlertSrv.IsAdmin(c, req.Applicant); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if isAdmin {
		req.CMDBs = []string{}
	}

	if getAlertMessageAbstractVo.EnableRules, getAlertMessageAbstractVo.DisableRules, err = api.AlertSrv.FindAlertRuleTypeCount(req.CMDBs); err != nil {
		log.Errorf("GetAlertMessageAbstract FindAlertRuleTypeCount error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if getAlertMessageAbstractVo.CurrentDay, err = api.AlertSrv.FindAlertMessageCountBySpecCond(&request.GetAlertMessageListRequest{
		CMDBs:     req.CMDBs,
		DBEnv:     req.DBEnv,
		StartTime: int(start.AddDate(0, 0, -1).Unix()),
		EndTime:   int(start.Unix()),
	}); err != nil {
		log.Errorf("GetAlertMessageAbstract FindAlertMessageCountByCond CurrentDay error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if getAlertMessageAbstractVo.Last10Min, err = api.AlertSrv.FindAlertMessageCountBySpecCond(&request.GetAlertMessageListRequest{
		CMDBs:     req.CMDBs,
		DBEnv:     req.DBEnv,
		StartTime: int(start.Add(-10 * time.Minute).Unix()),
		EndTime:   int(start.Unix()),
	}); err != nil {
		log.Errorf("GetAlertMessageAbstract FindAlertMessageCountByCond Last10Min error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if getAlertMessageAbstractVo.MuteNum, err = api.AlertSrv.FindAlertMessageCountBySpecCond(&request.GetAlertMessageListRequest{
		CMDBs:     req.CMDBs,
		DBEnv:     req.DBEnv,
		IsMute:    true,
		StartTime: int(time.Time{}.Unix()),
		EndTime:   int(time.Time{}.Unix()),
	}); err != nil {
		log.Errorf("GetAlertMessageAbstract FindAlertMessageCountByCond MuteNum error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, getAlertMessageAbstractVo, err)
}

func (api *Api) CreateAlertMessageMute(c *gin.Context) {
	var (
		err            error
		du             time.Duration
		muteConfigList []*alert.MuteConfigManager
		resp           = response.MuteResponse{}
		req            = &request.CreateMuteRequest{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	req.Applicant = api.GetOperatorEmail(c)

	if req.MuteTime == "forever" {
		req.MuteTime = MuteForever
	}

	if du, err = time.ParseDuration(req.MuteTime); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("time error, time=%s,err=%s", req.MuteTime, err.Error()))
		return
	}

	if muteConfigList, err = api.AlertSrv.FetchMuteRequestList(req.AlertIds, du, req.DBEnv, req.Applicant); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if resp.SuccessAlerts, resp.FailAlerts, err = api.AlertSrv.BatchCreateMute(muteConfigList); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, resp, nil)
}

func (api *Api) DeleteAlertMessageMute(c *gin.Context) {
	var (
		err  error
		resp = response.MuteResponse{}
		req  = &request.CancelMuteRequest{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	req.Applicant = api.GetOperatorEmail(c)
	if len(req.AlertIds) > BatchDealCount {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("batch delete mute count is %d", BatchDealCount))
		return
	}

	if resp.SuccessAlerts, resp.FailAlerts, err = api.AlertSrv.BatchCancelMute(req.AlertIds, req.Applicant, req.DBEnv); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, resp, nil)
}

func (api *Api) CreateAlertMessageAck(c *gin.Context) {
	var (
		err  error
		resp = response.AckResponse{}
		req  = &request.CreateAckRequest{}
	)

	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	req.Applicant = api.GetOperatorEmail(c)

	if resp.SuccessAlerts, resp.FailAlerts, err = api.AlertSrv.BatchCreateAck(req.AlertIds, req.Applicant); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	// record operator log
	if err = api.alertOpLog.Record(req.Applicant, strings.Join(req.AlertIds, ","), oplog.CREATE, oplog.CREATE_ACK, req.DBEnv, "", req); err != nil {
		log.Warningf("CreateAlertMessageAck alertOpLog error:%s", err.Error())
	}

	response.ToResponse(c, resp, nil)
}
