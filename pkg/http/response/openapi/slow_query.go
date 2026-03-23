package openapi

import (
	conf "smart-slowquery/conf/openapi"
	storeResp "smart-slowquery/pkg/store/response"

	"fmt"
)

type SlowQueryDBStatisticListVo struct {
	List      []*SlowSlowQueryDBStatisticVo `json:"dbs"`
	Count     int                           `json:"count"`
	Env       string                        `json:"environment"`
	StartTime int64                         `json:"start_time"`
	EndTime   int64                         `json:"end_time"`
}

type SlowSlowQueryDBStatisticVo struct {
	FingerID       string  `json:"finger_id"`
	FingerSQL      string  `json:"finger_sql"`
	DBName         string  `json:"database_name"`
	ClusterUUid    string  `json:"cluster_uuid"`
	TotalCount     int     `json:"total_count"`
	Role           string  `json:"role"`
	P80Time        float64 `json:"p80_time"`
	TotalQueryTime float64 `json:"total_query_time"`
	AvgQueryTime   float64 `json:"avg_query_time"`
	TotalLockTime  float64 `json:"total_lock_time"`
	DetailUrl      string  `json:"detail_url"`
}

func BuildSlowDBStatisticListVo(list *[]storeResp.SlowQueryDBStatistic, env string, startTime, endTime int64) *SlowQueryDBStatisticListVo {
	resp := &SlowQueryDBStatisticListVo{}

	for _, one := range *list {
		statisticVo := &SlowSlowQueryDBStatisticVo{
			FingerID:       one.FingerID,
			FingerSQL:      one.FingerSQL,
			Role:           one.Role,
			P80Time:        one.Top80Time,
			AvgQueryTime:   one.AvgQueryTime,
			DBName:         one.DBName,
			ClusterUUid:    one.ClusterUUID,
			TotalCount:     one.Count,
			TotalQueryTime: one.TotalQueryTime,
			TotalLockTime:  one.TotalLockTime,
			DetailUrl:      fmt.Sprintf(conf.GlobalConfig.DBDetailUrl, env, one.ClusterUUID, one.DBName, startTime, endTime) + "&fingerId=" + one.FingerID,
		}
		resp.List = append(resp.List, statisticVo)
	}
	resp.Count = len(resp.List)
	resp.Env = env
	resp.StartTime = startTime
	resp.EndTime = endTime
	return resp
}
