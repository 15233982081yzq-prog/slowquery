package http

import (
	timeUtil "smart-slowquery/internal/util/time"
	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"smart-slowquery/internal/model/report"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"

	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultOrderBy = "query_time"
	reportSwitch   = "open"
)

func (api *Api) DBSlowQueryRank(c *gin.Context) {
	var (
		err                    error
		top                    int
		dayStr, orderBy, dbEnv string
		toReport               bool
		day                    time.Time
		dailyRank              *report.DBSlowQueryRankDaily
	)

	toReport = c.DefaultQuery("to_report", "close") == reportSwitch
	dayStr = c.DefaultQuery("rank_day", "")
	dbEnv = c.DefaultQuery("db_env", "")

	if day, err = time.Parse(timeUtil.DayFormat, dayStr); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,day format invalid"))
		return
	}

	if orderBy = c.DefaultQuery("order_by", defaultOrderBy); len(orderBy) == 0 || orderBy != defaultOrderBy {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,order_by:%s invalid", orderBy))
		return
	}

	if top, err = strconv.Atoi(c.DefaultQuery("top", "10")); err != nil || top < 10 || top > 30 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,top:%d invalid", top))
		return
	}

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("rankSrv.DBSlowQueryRank", sysMetrics.GetStatus(err), time.Since(start))

	if dailyRank, err = api.reportAct.DBDailyRank(day, dbEnv, top); err != nil {
		log.Errorf("reportAct.DBDailyRank order_by:%s ,error:%s", orderBy, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	if dailyRank == nil {
		log.Warningf("reportAct.DBDailyRank order_by:%s ,rank is empty", orderBy)
		response.ToResponse(c, map[string]string{}, nil)
		return
	}

	log.Infof("DBSlowQueryRank rank_size:%d", len(dailyRank.Rank))

	if toReport {
		if err = api.emailSrv.SendDBRankReport(dailyRank); err != nil {
			log.Errorf("DBSlowQueryRank emailSrv SendDBRankReport error:%s", err.Error())
			response.ToResponse(c, map[string]string{}, err)
			return
		}
	}

	response.ToResponse(c, dailyRank, err)
}

func (api *Api) FingerSlowQueryRank(c *gin.Context) {
	var (
		err                    error
		top                    int
		toReport               bool
		dayStr, orderBy, dbEnv string
		day                    time.Time
		fingerRank             *report.FingerSlowQueryRankDaily
	)

	toReport = c.DefaultQuery("to_report", "close") == reportSwitch
	dayStr = c.DefaultQuery("rank_day", "")
	dbEnv = c.DefaultQuery("db_env", "")

	if day, err = time.Parse(timeUtil.DayFormat, dayStr); err != nil {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,day format invalid"))
		return
	}

	if orderBy = c.DefaultQuery("order_by", defaultOrderBy); len(orderBy) == 0 || orderBy != defaultOrderBy {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,order_by:%s invalid", orderBy))
		return
	}

	if top, err = strconv.Atoi(c.DefaultQuery("top", "10")); err != nil || top < 10 || top > 30 {
		response.ToResponse(c, map[string]string{}, fmt.Errorf("params error ,top:%d invalid", top))
		return
	}

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("rankSrv.FingerSlowQueryRank", sysMetrics.GetStatus(err), time.Since(start))
	if fingerRank, err = api.reportAct.FingerDailyRank(day, dbEnv, top); err != nil {
		log.Errorf("reportAct.FingerDailyRank order_by:%s ,error:%s", orderBy, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	if fingerRank == nil {
		log.Warningf("reportAct.FingerDailyRank order_by:%s ,rank is empty", orderBy)
		response.ToResponse(c, map[string]string{}, nil)
		return
	}

	log.Infof("FingerSlowQueryRank rank_size:%d", len(fingerRank.Rank))

	if toReport {
		if err = api.emailSrv.SendFingerRankReport(fingerRank); err != nil {
			log.Errorf("FingerSlowQueryRank emailSrv SendFingerRankReport error:%s", err.Error())
			response.ToResponse(c, map[string]string{}, err)
			return
		}
	}

	response.ToResponse(c, fingerRank, err)
}
