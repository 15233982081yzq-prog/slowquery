package http

import (
	openApiResp "smart-slowquery/pkg/http/response/openapi"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	storeResp "smart-slowquery/pkg/store/response"

	"smart-slowquery/internal/util/errors"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"

	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func (api *Api) DBStatistics(c *gin.Context) {
	var (
		err       error
		respList  []storeResp.SlowQueryDBStatistic
		storeList *[]storeResp.SlowQueryDBStatistic
	)

	req := &request.QueryDBStatistic{}
	if err = BindJsonParam(c, req); err != nil {
		log.Errorf("http SetRemoteDBPasswd param failed, error:%s \n", err.Error())
		response.ToAbortErrorResponse(c, errors.AnnotateParameterErrorf(err, "http request error"))
		return
	}

	if len(req.DBEnv) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,db_env is empty"))
		return
	}

	if len(req.DBs) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,db_names is empty"))
		return
	}

	if len(req.GetDBNames()) > MAXBatchSize {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,max support 20 db of one request"))
		return
	}

	if len(req.GetClusterUUids()) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,cluster_uuids is empty"))
		return
	}

	if len(req.GetClusterUUids()) > MAXBatchSize {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,max support 20 cluster of one request"))
		return
	}

	if req.StartTime > req.EndTime {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("end_time do not less than start_time "))
		return
	}

	if req.EndTime-req.StartTime > OVERTIME {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("Time range must be less than 72h "))
		return
	}

	if time.Now().Unix()-req.StartTime > OVERTIME {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("The start time cannot be lower than 72h of the current time "))
		return
	}

	log.Infof("query statistics param dbEnv:%s,dbNames:%v,clusterUUids:%v,startTime:%d,endTime:%d",
		req.DBEnv, req.GetDBNames(), req.GetClusterUUids(), req.StartTime, req.EndTime)

	start := time.Now()
	storeList, err = api.querySrv.GetSlowQueryStatistics(req.DBEnv, req.GetClusterUUids(), req.GetDBNames(), req.StartTime, req.EndTime)
	sysMetrics.CollectServiceMetrics("querySrv.GetSlowQueryStatistics", sysMetrics.GetStatus(err), time.Since(start))

	if err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	for i := 0; i < len(*storeList); i++ {
		if !req.ValidMapping((*storeList)[i].ClusterUUID, (*storeList)[i].DBName) {
			log.Warningf("ck result:%v not include request.env:%s", (*storeList)[i], req.DBEnv)
			continue
		}
		respList = append(respList, (*storeList)[i])
	}

	response.ToResponse(c, openApiResp.BuildSlowDBStatisticListVo(&respList, req.DBEnv, req.StartTime, req.EndTime), err)
}
