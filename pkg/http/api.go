package http

import (
	"fmt"
	"smart-slowquery/internal/util/errors"
	"smart-slowquery/internal/util/sys"
	"smart-slowquery/pkg/action"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/oplog"
	"smart-slowquery/pkg/service"
	"smart-slowquery/pkg/service/alert"
	"smart-slowquery/pkg/service/dbms"
	"smart-slowquery/pkg/service/optimize"
	"smart-slowquery/pkg/service/report/email"
	"strings"
	"time"

	timeUtil "smart-slowquery/internal/util/time"
	cmdbService "smart-slowquery/pkg/service/cmdb"
	rankService "smart-slowquery/pkg/service/statistics/daily"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"go.uber.org/atomic"
)

const (
	APIVersion      = "v1"
	OPENAPIVersion  = "v1"
	ALERTAPIVersion = "v1"
	OVERTIME        = 14 * timeUtil.OneDayBySecond
	MAXBatchSize    = 20

	MINMetricStep = 15 //最小数据抓取间隔15s
)

var MINMetricRange = int64(time.Minute.Seconds() * 10)

type Api struct {
	engine      *gin.Engine
	Running     *atomic.Uint64
	querySrv    *service.QueryService
	rankSrv     *rankService.SlowQueryRankService
	accessSrv   *service.AccessService
	dbmsSrv     *dbms.DBMetaService
	dataBaseSrv *cmdbService.DataBasesService
	cmdbSrv     *cmdbService.Service
	reportAct   *action.ReportRankAction
	emailSrv    *email.RankReportService
	optimizeSrv *optimize.Service
	AlertSrv    *alert.Service
	alertOpLog  *oplog.AlertOpLog
}

func NewPlatformApi(engine *gin.Engine, querySrv *service.QueryService, accessSrv *service.AccessService, dataBaseSrv *cmdbService.DataBasesService, dbmsSrv *dbms.DBMetaService, rankSrv *rankService.SlowQueryRankService, cmdbSrv *cmdbService.Service, emailSrv *email.RankReportService, reportAct *action.ReportRankAction, optimizeSrv *optimize.Service) (*Api, error) {
	api := &Api{
		engine:      engine,
		Running:     atomic.NewUint64(0),
		querySrv:    querySrv,
		rankSrv:     rankSrv,
		accessSrv:   accessSrv,
		dataBaseSrv: dataBaseSrv,
		dbmsSrv:     dbmsSrv,
		cmdbSrv:     cmdbSrv,
		emailSrv:    emailSrv,
		reportAct:   reportAct,
		optimizeSrv: optimizeSrv,
	}
	return api, nil
}

func (api *Api) AddPlatformApi(e *gin.Engine) {

	apiGroup := e.Group("/rds/smart/v1/api/")
	// 以下接口需要使用env路由
	{
		slowQueryGroup := apiGroup.Group(":server_env/slowquery")
		slowQueryGroup.GET("/db_metrics", api.DBMetrics)
		slowQueryGroup.GET("/query_list", api.QueryList)
		slowQueryGroup.GET("/query_detail", api.QueryDetail)
		slowQueryGroup.POST("/query_explain", api.QueryExplain)
		slowQueryGroup.GET("/client_trace", api.ClientTraceabilityByFinger)
		slowQueryGroup.GET("/statement_list", api.StatementsByFingerID)
		slowQueryGroup.GET("/db_daily_rank", api.DBSlowQueryRank)
		slowQueryGroup.GET("/finger_daily_rank", api.FingerSlowQueryRank)
	}
	{
		spaceGroup := apiGroup.Group(":server_env/cmdb")
		spaceGroup.GET("/services", api.GetServiceTree)
		spaceGroup.GET("/role", api.GetUserRole)
		spaceGroup.GET("/logic_dbs", api.GetLogicDBByService)
	}
	{
		platFromGroup := apiGroup.Group(":server_env/control")
		platFromGroup.POST("/server503", api.Server503)
		platFromGroup.POST("/server200", api.Server200)
		platFromGroup.GET("/switch", api.Switcher)
		platFromGroup.GET("/running", api.ServerRunning)
	}
	{
		platFromGroup := apiGroup.Group(":server_env/platform")
		platFromGroup.GET("/database/hosts", api.DataBaseHosts)
	}
	{
		accessGroup := apiGroup.Group("access")
		accessGroup.POST("/set_passwd", api.SetRemoteDBPasswd)
		accessGroup.GET("/get_passwd", api.GetRemoteDBPasswd)
	}
	{
		optGroup := apiGroup.Group(":server_env/platform")
		optGroup.POST("/optimize", api.Optimize)
	}
}

func NewOpenAPIApi(engine *gin.Engine, querySrv *service.QueryService) (*Api, error) {
	api := &Api{
		engine:   engine,
		Running:  atomic.NewUint64(0),
		querySrv: querySrv,
	}
	return api, nil
}

func (api *Api) AddOpenApi(e *gin.Engine, limiter *ratelimit.Bucket) {

	apiGroup := e.Group("/rds/smart/v1/openapi/")
	// 以下接口需要使用env路由

	slowQueryGroup := apiGroup.Group(":server_env/slowquery")
	slowQueryGroup.POST("/query_db_statistic", ratelimitMiddleware(limiter), api.DBStatistics)

}

func NewAlertAPIApi(engine *gin.Engine, alertSrv *alert.Service, alertOpLog *oplog.AlertOpLog, databaseSrv *cmdbService.DataBasesService) (*Api, error) {
	api := &Api{
		engine:      engine,
		AlertSrv:    alertSrv,
		alertOpLog:  alertOpLog,
		dataBaseSrv: databaseSrv,
	}
	return api, nil
}

func (api *Api) AddAlertApi(e *gin.Engine) {
	apiGroup := e.Group("/rds/smart/v1/alert_api/")

	// 以下接口需要使用env路由
	{
		ruleGroup := apiGroup.Group(":server_env/slowquery/rule")
		//报警规则
		ruleGroup.POST("/alert_rule", api.CreateAlertRule)
		ruleGroup.PUT("/alert_rule", api.UpdateAlertRule)
		ruleGroup.DELETE("/alert_rule", api.DeleteRules)
		ruleGroup.PATCH("/alert_rule", api.UpdateAlertRuleStatus)
		ruleGroup.PATCH("/alert_rules", api.BatchUpdateAlertRulesStatus)
		ruleGroup.POST("/alert_rules", api.GetAlertRule)
		ruleGroup.GET("/alert_rule_detail", api.GetAlertRuleDetail)

		//报警模版
		ruleGroup.GET("/alert_template", api.CreateAlertRuleTemplate)
	}
	{
		msgGroup := apiGroup.Group(":server_env/slowquery/message")
		//报警消息
		msgGroup.POST("/alert_message", api.GetAlertMessage)
		msgGroup.POST("/alert_message_abstract", api.GetAlertMessageAbstract)
		msgGroup.POST("/alert_message/mute", api.CreateAlertMessageMute)
		msgGroup.DELETE("/alert_message/mute", api.DeleteAlertMessageMute)
		msgGroup.PUT("/alert_message/ack", api.CreateAlertMessageAck)
	}
	{
		spaceGroup := apiGroup.Group(":server_env/space")
		//报警消息
		spaceGroup.GET("/logic_dbs", api.GetLogicDBByService)
	}
	{
		callbackGroup := apiGroup.Group(":server_env/slowquery/monitor")
		//monitor平台报警callback
		callbackGroup.POST("/alert_callback", api.AlertMessageCallback)
	}

}

func (api *Api) GetOperatorEmail(c *gin.Context) string {
	email, ok := c.Get(sys.ShopeeSpaceEmail)
	if !ok {
		return ""
	}
	return email.(string)
}

func (api *Api) GetToken(c *gin.Context) (token string) {
	if author := c.GetHeader("Authorization"); strings.Contains(author, "Bearer") {
		if authorArray := strings.Split(author, " "); len(authorArray) > 1 {
			token = authorArray[1]
		}
	}
	return token
}

func BindJsonParam(c *gin.Context, req interface{}) (err error) {
	return c.BindJSON(req)
}

// Middleware function to apply rate limiting
func ratelimitMiddleware(bucket *ratelimit.Bucket) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			response.ToAbortWithErrorCodeResponse(c, errors.NewError(429, "Too Many Requests:", fmt.Sprintf("ratelimit capacity:%d/second", bucket.Capacity())))
			return
		}
		c.Next()
	}
}
