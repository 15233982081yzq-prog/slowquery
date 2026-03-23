package clickhouse

import (
	sysMetrics "smart-slowquery/pkg/metrics/analyzer"

	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/request"

	"time"
)

func (cli *Client) batchPut(batch []*request.SlowQueryLog) (err error) {
	start := time.Now()
	defer func() {
		sysMetrics.CollectStoreMetrics("batchPut", sysMetrics.GetStatus(err), time.Since(start))
		sysMetrics.CollectStoreBatchCounter("batchPut", sysMetrics.GetStatus(err), len(batch))
	}()

	if len(batch) == 0 {
		log.Warning("ck client  batchPut batch is empty!")
		return nil
	}

	if err = cli.db.CreateInBatches(batch, len(batch)).Error; err != nil {
		log.Errorf("ck client batchPut create error:%s", err.Error())
		return err
	}
	log.Infof("batch put table:%s batch size:%d", batch[0].TableName(), len(batch))
	return nil
}

func (cli *Client) CreateAlertOperatorLog(req *request.AlertOperatorLog) (err error) {
	return cli.db.Create(req).Error
}

func (cli *Client) CreateAlertMessage(req *request.AlertMessage) (err error) {
	return cli.db.Create(req).Error
}

func (cli *Client) UpdateAlertMessage(req *request.AlertMessage) (err error) {
	tx := cli.db.Exec("ALTER TABLE "+req.LocalTableName()+" ON CLUSTER "+req.ClusterName()+" UPDATE update_time = ?, last_alert_time = ?, resolve_time=?, alert_count=?,alert_status=? WHERE monitor_alert_id=?", req.UpdateTime, req.LastAlertTime, req.ResolveTime, req.AlertCount, req.Status, req.MonitorAlertID)
	return tx.Error
}

func (cli *Client) UpdateAlertMessageStatus(req *request.AlertMessage) error {
	tx := cli.db.Exec("ALTER TABLE "+req.LocalTableName()+" ON CLUSTER "+req.ClusterName()+" UPDATE alert_status = ?, update_time = ? where monitor_alert_id=?", req.Status, req.UpdateTime, req.MonitorAlertID)
	return tx.Error
}

func (cli *Client) CreateAlertMute(req *request.AlertMute) (err error) {
	return cli.db.Create(req).Error
}

func (cli *Client) DeleteAlertMute(req *request.AlertMute) (err error) {
	tx := cli.db.Exec("ALTER TABLE "+req.LocalTableName()+" ON CLUSTER "+req.ClusterName()+" DELETE where monitor_mute_id=?", req.MonitorAlertID)
	return tx.Error
}

func (cli *Client) UpdateMuteStatus(req *request.AlertMute) (err error) {
	tx := cli.db.Exec("ALTER TABLE "+req.LocalTableName()+" ON CLUSTER "+req.ClusterName()+" UPDATE status = ?,update_time = ? WHERE monitor_mute_id = ?", req.Status, req.UpdateTime, req.MonitorMuteID)
	return tx.Error
}

func (cli *Client) BatchUpdateMuteStatusToTTL(req *request.AlertMute, dstStatus string, ts int64) (err error) {
	tx := cli.db.Exec("ALTER TABLE "+req.LocalTableName()+" ON CLUSTER "+req.ClusterName()+" UPDATE status=?,update_time = ? WHERE status = ? and end_time < ?", dstStatus, req.UpdateTime, req.Status, ts)
	return tx.Error
}
