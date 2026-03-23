package http

import (
	requestMeta "smart-slowquery/pkg/store/request"
	storeResp "smart-slowquery/pkg/store/response"

	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"

	"encoding/json"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

/**
 * 报警平台回调接口
 * 用于同步monitor服务慢日志报警信息到slow query alert平台
 */

func (api *Api) AlertMessageCallback(c *gin.Context) {
	var (
		err           error
		bytes         []byte
		alertMessages []*storeResp.AlertMessage
		alertList     []*request.AlertInfo
		req           = &request.AlertCallBackRequest{}
	)
	if err = BindJsonParam(c, req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if len(req.AlertList) > 0 {
		req.AlertList = req.AlertList[:len(req.AlertList)-1]
	}

	if bytes, err = json.Marshal(req); err != nil {
		log.Errorf("Marshal AlertCallBackRequest error:%s", err.Error())
	}
	log.Infof("AlertMessageCallback req data:%s", string(bytes))

	if req.FromSlowQuery() {
		if alertMessages, err = api.AlertSrv.GetAlertMessageByRuleIDAndMessageStatus(req.AlertRuleUuid, requestMeta.AlertPending); err != nil {
			log.Errorf("AlertMessageCallback ruleUUID:%s get alertMessage error:%s", req.AlertRuleUuid, err.Error())
			response.ToResponse(c, map[string]string{}, err)
			return
		}

		if len(alertMessages) == 0 {
			log.Warningf("AlertMessageCallback ruleUUID:%s get alertMessage is empty", req.AlertRuleUuid)
			response.ToResponse(c, map[string]string{}, nil)
			return
		}

		//去拼凑消息体
		for _, msg := range alertMessages {
			alertId, err := strconv.ParseInt(msg.MonitorAlertID, 10, 64)
			if err != nil {
				log.Warningf("AlertMessageCallback alertMessage alertID do strconv error:%s", err.Error())
				return
			}
			alertList = append(alertList, &request.AlertInfo{
				AlertId:     alertId,
				StartTime:   msg.StartTime,
				AlertRuleId: req.AlertRuleUuid,
				Status:      requestMeta.AlertClosed,
			})
		}
		req.AlertList = alertList
	}

	if len(req.AlertList) == 0 {
		response.ToResponse(c, map[string]string{}, errors.New("empty alert list"))
		return
	}

	if err = api.AlertSrv.AlertMessageCallback(req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	response.ToResponse(c, map[string]string{}, nil)
}
