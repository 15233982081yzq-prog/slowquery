package clickhouse

import (
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"

	"time"
)

func (cli *Client) GetFingerIdCountByTimeRange(req *request.SlowQueryDistinctFinger) (count int64, err error) {
	tx := cli.db.Select("count(DISTINCT(finger_id))").Table(req.GetTableName())
	if len(req.StartTime) > 0 && len(req.EndTime) > 0 {
		tx.Where("create_time >= ? and create_time <= ? ", req.StartTime, req.EndTime)
	}
	if err = tx.Find(&count).Error; err != nil {
		log.Errorf("ck client GetFingerIdCount error:%s", err.Error())
	}
	return count, err
}

func (cli *Client) GetFingerIdList(req *request.SlowQueryDistinctFinger) (resp []*response.FingerIDWithEarliestAppearTime, err error) {
	var start = time.Now()

	defer log.Infof("ck GetFingerIdList cost:%s, start:%d,end:%d", time.Since(start).String(), req.Offset, req.Offset+req.Limit)
	tx := cli.db.Select("finger_id,min(log_time) as earliest_appear_time").Table(req.GetTableName()).
		Where("create_time >= ? and create_time <= ?", req.StartTime, req.EndTime).Group("finger_id").Order("finger_id").Offset(req.Offset).Limit(req.Limit)

	if err = tx.Find(&resp).Error; err != nil {
		log.Errorf("ck client GetFingerIdWithLastSeenTimeList error:%s", err.Error())
	}
	return resp, err
}

func (cli *Client) GetDBNameList() (resp []string, err error) {

	tx := cli.db.Select("DISTINCT(database_name)").Table("slow_query_log_all_rand")

	if err = tx.Find(&resp).Error; err != nil {
		log.Errorf("ck client GetDBNameList error:%s", err.Error())
	}
	return resp, err
}
