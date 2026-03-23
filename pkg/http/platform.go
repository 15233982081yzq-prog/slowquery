package http

import (
	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service"

	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (api *Api) DataBaseHosts(c *gin.Context) {
	var (
		err                error
		hosts              []string
		startTime, endTime int64
	)

	dbName := c.DefaultQuery("db_name", "")
	dbEnv := c.DefaultQuery("db_env", "")
	if startTime, err = strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,start_time not number type"))
		return
	}

	if endTime, err = strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,end_time not number type"))
		return
	}

	log.Infof("DataBaseHosts dbName:%s,dbEnv:%s", dbName, dbEnv)

	start := time.Now()
	hosts, err = api.querySrv.GetInstanceHosts(dbName, dbEnv, startTime, endTime)
	sysMetrics.CollectServiceMetrics("querySrv.GetInstanceHosts", sysMetrics.GetStatus(err), time.Since(start))

	if err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("platform InstanceHosts error:%s", err.Error()))
		return
	}

	response.ToResponse(c, &response.InstanceHosts{Hosts: hosts}, err)
}

func (api *Api) CoreDBClusters(c *gin.Context) {
	coreClusters, err := service.GetCoreClusters()
	response.ToResponse(c, coreClusters, err)
}
