package http

import (
	envUtil "smart-slowquery/internal/util/env"
	monitorUtil "smart-slowquery/internal/util/monitor"
	stringUtil "smart-slowquery/internal/util/string"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	storeReq "smart-slowquery/pkg/store/request"

	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"

	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (api *Api) DBMetrics(c *gin.Context) {
	var (
		err                                    error
		startTime, endTime                     int64
		metricStep                             int
		dbEnv, metricName, metricType, dbNames string
		appearType                             *storeReq.AppearType
		md                                     *monitorUtil.MetricData
	)

	dbEnv = c.DefaultQuery("db_env", envUtil.ServerLiveEnv)

	if appearType = request.AppearedMapping[c.DefaultQuery("appear_type", "all")]; appearType == nil || appearType.IsOriginal() {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("param error ,appear_type is invalid"))
		return
	}

	if dbNames = c.DefaultQuery("db_names", ""); len(dbNames) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,db_names is empty"))
		return
	}

	if len(stringUtil.Split(dbNames, ",")) > 10 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbNames over limit 10 databases"))
		return
	}

	if metricName = c.DefaultQuery("metric_name", ""); len(metricName) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,metric_name is empty"))
		return
	}

	if metricType = c.DefaultQuery("metric_type", ""); len(metricType) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,metric_type is invalid"))
		return
	}

	if metricStep, err = strconv.Atoi(c.DefaultQuery("metric_step", "0")); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,metric_step need number value"))
		return
	}

	if startTime, err = strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,start_time not number type"))
		return
	}

	if endTime, err = strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,end_time not number type"))
		return
	}

	if startTime > endTime {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("end_time do not less than start_time "))
		return
	}

	if endTime-startTime > OVERTIME {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("Time range must be less than 72h "))
		return
	}

	if time.Now().Unix()-startTime > OVERTIME {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("The start time cannot be lower than 72h of the current time "))
		return
	}

	if endTime-startTime <= MINMetricRange {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("Time range must be more than 10m "))
		return
	}

	if metricStep < MINMetricStep {
		metricStep = MINMetricStep
	}

	log.Infof("DBMetrics param db_names:%s,db_env:%s,appearType:%s,metric_name:%s,metric_type:%s,metric_step:%d,startTime:%d,endTime:%d",
		dbNames, dbEnv, appearType.GetName(), metricName, metricType, metricStep, startTime, endTime)

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("MonitorUtil.GetMetric", sysMetrics.GetStatus(err), time.Since(start))
	if md, err = monitorUtil.GetMetric(strings.Split(dbNames, ","), metricName, metricType, metricStep, startTime, endTime, appearType); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	response.ToResponse(c, md, err)
}
