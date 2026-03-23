package alert

import (
	httpReq "smart-slowquery/pkg/http/request"
	httpResp "smart-slowquery/pkg/http/response"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	storeReq "smart-slowquery/pkg/store/request"
	storeResp "smart-slowquery/pkg/store/response"

	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/oplog"

	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MuteTimeFormat = "2006-01-02T15:04:05Z"
	MuteOnce       = "once"
	MuteDesc       = "Auto create by slow query, alert_id=%v, mute_uuid=%v, alert_rule_uuid=%s, channel_uuid=%s"
	MuteVersion    = "mon.alertRuleMute.muteFilter/v1"

	MuteMutedStatus   = "muted"
	MuteUnMutedStatus = "unmuted"
	MuteTTLStatus     = "ttl"
)

func (s *Service) FetchMuteRequestList(ids []string, muteDuration time.Duration, env, user string) ([]*MuteConfigManager, error) {
	var (
		err              error
		resp             = make([]*MuteConfigManager, 0)
		alertMessageList []*storeResp.AlertMessage
		start            = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("alertSrv.FetchMuteRequestList", sysMetrics.GetStatus(err), time.Since(start))

	if alertMessageList, err = s.ck.GetAlertMessageByAlertIDs(ids); err != nil {
		return nil, err
	}

	now := time.Now()
	for _, msg := range alertMessageList {
		if msg.Status != storeReq.AlertPending {
			log.Errorf("alertSrv FetchMuteRequestList this alert has been %s, need`t to mute, monitor_alert_id:%s", msg.Status, msg.MonitorAlertID)
			return nil, fmt.Errorf("this alert has been %s, need`t to mute, monitor_alert_id:%s", msg.Status, msg.MonitorAlertID)
		}
		labelInfo := httpReq.AlertLabel{}
		if e := json.Unmarshal([]byte(msg.LabelInfo), &labelInfo); e != nil {
			log.Errorf("alert label json to struct error, json:%s", msg.LabelInfo)
			continue
		}

		mcm := &MuteConfigManager{
			Env:           env,
			MuteUuid:      uuid.New().String(),
			AlertId:       msg.MonitorAlertID,
			AlertRuleUuid: msg.AlertRuleUUID,
			Applicant:     user,
			ChannelUuid:   msg.ChannelUUID,
			MuteRange:     muteDuration.String(),
			MuteTimeRanges: []MuteTimeRange{
				{
					StartTime: now.Add(-time.Minute).UTC().Format(MuteTimeFormat),
					StartUnix: now.Add(-time.Minute).UTC().Unix(),
					EndTime:   now.Add(muteDuration).UTC().Format(MuteTimeFormat),
					EndUnix:   now.Add(muteDuration).UTC().Unix(),
					Type:      MuteOnce,
				},
			},
			FilterList: genLabelFilter(labelInfo),
		}
		resp = append(resp, mcm)
	}
	return resp, nil
}

func genLabelFilter(labelInfo httpReq.AlertLabel) (res []Filter) {
	if labelInfo.AlertGroup != "" {
		res = append(res, Filter{
			Key:      "alertgroup",
			Operator: "=",
			Value:    labelInfo.AlertGroup,
		})
	}

	if labelInfo.AlertName != "" {
		res = append(res, Filter{
			Key:      "alertname",
			Operator: "=",
			Value:    labelInfo.AlertName,
		})
	}

	if labelInfo.DatabaseName != "" {
		res = append(res, Filter{
			Key:      "database_name",
			Operator: "=",
			Value:    labelInfo.DatabaseName,
		})
	}

	if labelInfo.CMDB != "" {
		res = append(res, Filter{
			Key:      "cmdb",
			Operator: "=",
			Value:    labelInfo.CMDB,
		})
	}
	if labelInfo.MRuleID != "" {
		res = append(res, Filter{
			Key:      "m_rule_id",
			Operator: "=",
			Value:    labelInfo.MRuleID,
		})
	}
	return res
}

func (s *Service) BatchCreateMute(muteManagerList []*MuteConfigManager) (succList []string, failList []httpResp.FailInfo, err error) {
	var (
		wg    sync.WaitGroup
		mux   sync.Mutex
		start = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("alertSrv.BatchCreateMute", sysMetrics.GetStatus(err), time.Since(start))

	for _, mm := range muteManagerList {
		wg.Add(1)
		go func(mcm *MuteConfigManager) {
			_, e := s.createMute(mcm)
			mux.Lock()
			if e != nil {
				log.Errorf("BatchCreateMute createMute error:%s", e.Error())
				failList = append(failList, httpResp.FailInfo{
					UUID:   mcm.AlertId,
					ErrMsg: e.Error(),
				})
			} else {
				succList = append(succList, mcm.AlertId)
			}
			mux.Unlock()
			wg.Done()
		}(mm)
	}
	wg.Wait()
	return succList, failList, err
}

func (s *Service) BatchCancelMute(ids []string, user, env string) (succList []string, failList []httpResp.FailInfo, err error) {
	var (
		wg               sync.WaitGroup
		mux              sync.Mutex
		alertMessageList []*storeResp.AlertMessage
		start            = time.Now()
	)
	defer sysMetrics.CollectStoreMetrics("alertSrv.BatchCancelMute", sysMetrics.GetStatus(err), time.Since(start))

	if alertMessageList, err = s.ck.GetAlertMessageByAlertIDs(ids); err != nil {
		return nil, nil, err
	}

	for _, mm := range alertMessageList {
		wg.Add(1)
		go func(alertId, user string) {
			e := s.cancelMute(alertId, user, env)
			mux.Lock()
			if e != nil {
				log.Errorf("BatchCancelMute cancelMute error:%s", e.Error())
				failList = append(failList, httpResp.FailInfo{
					UUID:   alertId,
					ErrMsg: e.Error(),
				})
			} else {
				succList = append(succList, alertId)
			}
			wg.Done()
			mux.Unlock()
		}(mm.MonitorAlertID, user)
	}
	wg.Wait()
	return succList, failList, err
}

func (s *Service) BatchCreateAck(ids []string, user string) (succList []string, failList []httpResp.FailInfo, err error) {
	var (
		wg    sync.WaitGroup
		mux   sync.Mutex
		start = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("alertSrv.BatchCreateAck", sysMetrics.GetStatus(err), time.Since(start))

	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			e := s.createAck(id)
			mux.Lock()
			if e != nil {
				log.Errorf("BatchCreateAck createAck error:%s", e.Error())
				failList = append(failList, httpResp.FailInfo{
					UUID:   id,
					ErrMsg: e.Error(),
				})
			} else {
				succList = append(succList, id)
			}
			mux.Unlock()
			wg.Done()
		}(id)
	}
	wg.Wait()
	return succList, failList, err
}

func (s *Service) createMute(mcm *MuteConfigManager) (muteUuid string, err error) {
	var (
		filterBytes  []byte
		muteInfoList []*storeResp.AlertMute
		resp         *CreateAlertRuleMuteConfigResponse
		start        = time.Now()
	)

	defer sysMetrics.CollectStoreMetrics("createMute", sysMetrics.GetStatus(err), time.Since(start))

	if muteInfoList, err = s.ck.GetAlertMute(mcm.AlertId); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	for _, muteInfo := range muteInfoList {
		if muteInfo != nil && muteInfo.Status == MuteMutedStatus {
			return "", fmt.Errorf("this alert has been mute by %s, alert_id:%s", muteInfo.Creator, mcm.AlertId)
		}
	}

	req := &CreateAlertRuleMuteConfigRequest{
		DisplayName:  fmt.Sprintf("slow-query-mute-rule-%s", mcm.MuteUuid),
		MuteDesc:     fmt.Sprintf(MuteDesc, mcm.AlertId, mcm.MuteUuid, mcm.AlertRuleUuid, mcm.ChannelUuid),
		MuteLabelOrg: LabelOrg,
		ProjectName:  ProjectName,
		MuteFilter: MuteFilter{
			Version: MuteVersion,
			Spec: struct {
				FilterList []Filter `json:"filter,omitempty"`
			}(
				struct{ FilterList []Filter }{FilterList: mcm.FilterList}),
		},
		MuteTimeRange: mcm.MuteTimeRanges,
		IsDisabled:    0,
	}
	if resp, err = s.monitorClient.CreateAlertRuleMuteConfig(req); err != nil {
		log.Warningf("create alert rule failed:%v", err)
		return "", err
	}

	if filterBytes, err = json.Marshal(mcm.FilterList); err != nil {
		return "", err
	}

	muteRequest := storeReq.BuildAlertMute(
		resp.MuteId,
		mcm.Env,
		req.DisplayName,
		mcm.AlertRuleUuid,
		mcm.AlertId,
		MuteMutedStatus,
		string(filterBytes),
		mcm.Applicant,
		uint64(mcm.MuteTimeRanges[0].StartUnix),
		uint64(mcm.MuteTimeRanges[0].EndUnix))

	if err = s.ck.CreateAlertMute(muteRequest); err != nil {
		return "", err
	}

	// record operator log
	if err = s.alertOpLog.Record(mcm.Applicant, mcm.AlertId, oplog.CREATE, oplog.CREATE_MUTE, mcm.Env, "", muteRequest); err != nil {
		log.Warningf("record create mute log err:%s", err.Error())
	}

	return resp.MuteId, nil
}

func (s *Service) cancelMute(alertID, user, env string) (err error) {
	var (
		currentMute  *storeResp.AlertMute
		muteInfoList []*storeResp.AlertMute
		resp         *GetSingleAlertRuleMuteConfigResponse
		start        = time.Now()
	)
	defer sysMetrics.CollectStoreMetrics("cancelMute", sysMetrics.GetStatus(err), time.Since(start))

	if muteInfoList, err = s.ck.GetAlertMute(alertID); err != nil {
		return err
	}

	for _, muteInfo := range muteInfoList {
		if muteInfo.Status == MuteMutedStatus {
			currentMute = muteInfo
			break
		}
	}

	if resp, err = s.monitorClient.GetSingleAlertRuleMuteConfig(currentMute.MonitorMuteID); err != nil {
		return err
	}

	req := &UpdateAlertRuleMuteConfigRequest{
		MuteConfig: RuleMuteConfig{
			IsDisabled:  1,
			DataVersion: resp.DataVersion,
		},
		UpdateMask: UpdateMask{Paths: []string{"is_disabled"}},
	}

	if _, err = s.monitorClient.UpdateSingleAlertRuleMuteConfig(currentMute.MonitorMuteID, req); err != nil {
		log.Warningf("create alert rule failed:%v", err)
		return err
	}

	// 更新软状态
	if err = s.ck.UpdateMuteStatus(currentMute.MonitorMuteID, MuteUnMutedStatus); err != nil {
		return err
	}
	// record operator log
	if err = s.alertOpLog.Record(user, alertID, oplog.DELETE, oplog.DELETE_MUTE, env, currentMute, ""); err != nil {
		log.Warningf("cancelMute alertOpLog error:%s", err.Error())
	}
	return nil
}

func (s *Service) createAck(alertID string) (err error) {
	start := time.Now()
	defer sysMetrics.CollectStoreMetrics("createAck", sysMetrics.GetStatus(err), time.Since(start))

	err = s.monitorClient.CreateAck(alertID)
	return
}
