package http

import (
	stringUtil "smart-slowquery/internal/util/string"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	storeReq "smart-slowquery/pkg/store/request"
	storeResp "smart-slowquery/pkg/store/response"

	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"

	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	orderByItem = map[string]bool{
		"count":      true,
		"total_time": true,
		"avg_time":   true,
	}
)

func (api *Api) QueryList(c *gin.Context) {
	var (
		err                                                        error
		page, pageSize, limit, offset, total                       int
		startTime, endTime                                         int64
		orderBy, cmdb, dbEnv, clusterUUIDs, dbNames, instanceHosts string
		appearType                                                 *storeReq.AppearType
		querys                                                     *[]storeResp.SlowQuery
	)

	if appearType = request.AppearedMapping[c.DefaultQuery("appear_type", "all")]; appearType == nil || appearType.IsOriginal() {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("param error ,appear is invalid"))
		return
	}

	if page, err = strconv.Atoi(c.Query("page")); err != nil || page <= 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,no page message"))
		return
	}

	if pageSize, err = strconv.Atoi(c.Query("page_size")); err != nil || pageSize <= 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,no page_size message"))
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

	limit = pageSize
	offset = pageSize * (page - 1)
	orderBy = c.DefaultQuery("order_by", "count")
	cmdb = c.DefaultQuery("cmdb", "")
	dbEnv = c.DefaultQuery("db_env", "")
	dbNames = c.DefaultQuery("db_names", "")
	clusterUUIDs = c.DefaultQuery("cluster_uuids", "")
	instanceHosts = c.DefaultQuery("instance_hosts", "")

	if len(dbNames) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbNames is empty"))
		return
	}

	if len(stringUtil.Split(dbNames, ",")) > 10 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbNames over limit 10 databases"))
		return
	}

	if len(dbEnv) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbEnv is empty"))
		return
	}

	if len(clusterUUIDs) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,clusterUUIDs is empty"))
		return
	}

	if !orderByItem[orderBy] {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("orderBy must optional : count | total_time | avg_time"))
		return
	}

	log.Infof("query list param page:%d,pageSize:%d,orderBy:%s,cmdb:%s,"+
		"dbEnv:%s,dbNames:%s,instanceHosts:%s,startTime:%d,endTime:%d",
		page, pageSize, orderBy, cmdb, dbEnv, dbNames, instanceHosts, startTime, endTime)

	start := time.Now()
	querys, total, err = api.querySrv.GetSlowQueryList(dbEnv, orderBy, stringUtil.Split(clusterUUIDs, ","), stringUtil.Split(dbNames, ","), stringUtil.Split(instanceHosts, ","), limit, offset, startTime, endTime, appearType)
	sysMetrics.CollectServiceMetrics("querySrv.GetSlowQueryList", sysMetrics.GetStatus(err), time.Since(start))

	if err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	// 检查出现的数据，是否在物化视图中，如果存在，则是新出现的指纹，反之则不是
	fingerList := make([]string, 0)
	for _, result := range *querys {
		fingerList = append(fingerList, result.FingerID)
	}
	// in方式查询
	mmap, _ := api.querySrv.FindLast7Days(fingerList)

	response.ToResponse(c, response.BuildSlowQueryListVo(querys, total, pageSize, mmap), err)
}

func (api *Api) QueryDetail(c *gin.Context) {
	var (
		err                error
		clientUser         string
		startTime, endTime int64
		users              []string
		appearType         *storeReq.AppearType
		stmt               *storeResp.QueryStatement
	)

	fingerID := c.DefaultQuery("finger_id", "")
	dbName := c.DefaultQuery("db_name", "")
	dbEnv := c.DefaultQuery("db_env", "")
	clusterUUID := c.DefaultQuery("cluster_uuid", "")
	instanceHosts := c.DefaultQuery("instance_hosts", "")

	if appearType = request.AppearedMapping[c.DefaultQuery("appear_type", "")]; appearType == nil || appearType.IsAll() {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("param error ,appear_type is invalid"))
		return
	}

	if len(fingerID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,fingerID is empty"))
		return
	}

	if len(dbName) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbName is empty"))
		return
	}

	if len(dbEnv) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbEnv is empty"))
		return
	}

	if len(clusterUUID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,clusterUUID is empty"))
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

	log.Infof("query detail param finger_id:%s,db_name:%s,db_env:%s,instance_hosts:%s,start_time:%d,end_time:%d", fingerID, dbName, dbEnv, instanceHosts, startTime, endTime)

	start := time.Now()
	stmt, users, err = api.querySrv.GetQueryDetail(dbName, dbEnv, fingerID, stringUtil.Split(instanceHosts, ","), startTime, endTime, appearType)
	sysMetrics.CollectServiceMetrics("querySrv.GetQueryDetail", sysMetrics.GetStatus(err), time.Since(start))
	if err != nil {
		log.Errorf("service.GetQueryDetail error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if len(users) > 0 {
		clientUser = users[0]
	}
	response.ToResponse(c, response.BuildSlowQueryDetailVo(stmt, dbName, clientUser), err)
}

func (api *Api) ClientTraceabilityByFinger(c *gin.Context) {
	var (
		err                error
		stats              *[]storeResp.ClientHostStats
		startTime, endTime int64
	)

	fingerID := c.DefaultQuery("finger_id", "")
	dbName := c.DefaultQuery("db_name", "")
	dbEnv := c.DefaultQuery("db_env", "")
	clusterUUID := c.DefaultQuery("cluster_uuid", "")
	instanceHosts := c.DefaultQuery("instance_hosts", "")

	if len(fingerID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,fingerID is empty"))
		return
	}

	if len(dbName) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbName is empty"))
		return
	}

	if len(dbEnv) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbEnv is empty"))
		return
	}

	if len(clusterUUID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,clusterUUID is empty"))
		return
	}

	if len(clusterUUID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,clusterUUID is empty"))
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

	log.Infof("queryFinger Client Traceability param finger_id:%s,db_name:%s,db_env:%s,instance_hosts:%s,start_time:%d,end_time:%d", fingerID, dbName, dbEnv, instanceHosts, startTime, endTime)

	start := time.Now()
	stats, err = api.querySrv.GetClientTraceability(dbName, dbEnv, fingerID, stringUtil.Split(instanceHosts, ","), startTime, endTime)
	sysMetrics.CollectServiceMetrics("querySrv.QueryFingerClientTraceability", sysMetrics.GetStatus(err), time.Since(start))

	if err != nil {
		log.Errorf("service.GetClientTraceability error:%s", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, response.BuildFingerClientTraceabilityVo(fingerID, stats), err)
}

func (api *Api) StatementsByFingerID(c *gin.Context) {
	var (
		err                                  error
		stmts                                *[]storeResp.QueryStatement
		page, pageSize, limit, offset, count int
		startTime, endTime                   int64
	)

	fingerID := c.DefaultQuery("finger_id", "")
	dbName := c.DefaultQuery("db_name", "")
	dbEnv := c.DefaultQuery("db_env", "")
	clusterUUID := c.DefaultQuery("cluster_uuid", "")
	instanceHosts := c.DefaultQuery("instance_hosts", "")
	orderBy := c.DefaultQuery("order_by", "query_time")

	if page, err = strconv.Atoi(c.Query("page")); err != nil || page <= 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,no page message"))
		return
	}

	if pageSize, err = strconv.Atoi(c.Query("page_size")); err != nil || pageSize <= 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,no page_size message"))
		return
	}

	if len(fingerID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,fingerID is empty"))
		return
	}

	if len(dbName) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbName is empty"))
		return
	}

	if len(dbEnv) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,dbEnv is empty"))
		return
	}

	if len(clusterUUID) == 0 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,clusterUUID is empty"))
		return
	}

	if orderBy != "query_time" && orderBy != "lock_time" {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,orderby value:%s is invalid,only supports query_time and lock_time fields", orderBy))
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

	log.Infof("QueryStatementsByFingerID param finger_id:%s,db_name:%s,db_env:%s,instance_hosts:%s,start_time:%d,end_time:%d", fingerID, dbName, dbEnv, instanceHosts, startTime, endTime)

	limit = pageSize
	offset = pageSize * (page - 1)
	start := time.Now()
	stmts, count, err = api.querySrv.GetQueryStatements(dbName, dbEnv, fingerID, orderBy, stringUtil.Split(instanceHosts, ","), startTime, endTime, offset, limit)
	sysMetrics.CollectServiceMetrics("querySrv.StatementsByFingerID", sysMetrics.GetStatus(err), time.Since(start))

	if err != nil {
		log.Errorf("service.StatementsByFingerID fingerID:%s, database_name:%s, error:%s", fingerID, dbName, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, response.BuildFingerStatementsVo(fingerID, count, pageSize, stmts), err)
}
