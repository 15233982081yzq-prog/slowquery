package oplog

import (
	"smart-slowquery/pkg/log"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/request"
	"time"

	"encoding/json"
)

const (
	CREATE = "create"
	UPDATE = "update"
	DELETE = "delete"

	BATCH_CREATE_ALERT        = "batch_create_alert"
	CREATE_ALERT              = "create_alert"
	UPDATE_ALERT              = "update_alert"
	UPDATE_ALERT_STATUS       = "update_alert_status"
	BATCH_UPDATE_ALERT_STATUS = "batch_update_alert_status"
	DELETE_ALERT              = "delete_alert"
	BATCH_DELETE_ALERT        = "batch_delete_alert"
	CREATE_MUTE               = "create_mute"
	DELETE_MUTE               = "delete_mute"
	CREATE_ACK                = "create_ack"

	MonitorPlatform        = "monitor_platform"
	SlowQueryAlertPlatform = "slow_query_alert_platform"
)

type AlertOpLog struct {
	ck store.CKStore
}

func NewAlertOpLog(ckStore store.CKStore) *AlertOpLog {
	return &AlertOpLog{
		ck: ckStore,
	}
}

func (oplog *AlertOpLog) Record(operator, actionID, actionType, actionName, env string, oldSetting, newSetting interface{}) error {
	log.Infof("operatorLog Commit operator:%s,actionID:%s,actionType:%s,actionName:%s,env:%s", operator, actionID, actionType, actionName, env)
	var (
		oldSettingBs, newSettingBs []byte
		err                        error
		start                      = time.Now()
	)

	defer func() {
		sysMetrics.CollectServiceMetrics("alertOpLog.Record", sysMetrics.GetStatus(err), time.Since(start))
	}()

	oldSettingBs, err = json.Marshal(oldSetting)
	if err != nil {
		return err
	}
	newSettingBs, err = json.Marshal(newSetting)
	if err != nil {
		return err
	}
	return oplog.ck.CreateAlertOperatorLog(request.BuildAlertOperatorLog(operator, actionID, actionType, actionName, env, string(oldSettingBs), string(newSettingBs)))
}
