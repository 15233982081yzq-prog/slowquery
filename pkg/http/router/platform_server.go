package router

import (
	conf "smart-slowquery/conf/platform"
	"smart-slowquery/internal/util/monitor"
	sysHttp "smart-slowquery/pkg/http"
	cmdbService "smart-slowquery/pkg/service/cmdb"
	rankService "smart-slowquery/pkg/service/statistics/daily"
	"smart-slowquery/pkg/service/uic"
	"smart-slowquery/pkg/store/clickhouse"
	thirdAdvisor "smart-slowquery/thrid-party/soar-dev/advisor"
	thirdCommon "smart-slowquery/thrid-party/soar-dev/common"

	"smart-slowquery/internal/util/sys"
	"smart-slowquery/pkg/action"
	"smart-slowquery/pkg/http/filter"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service"
	"smart-slowquery/pkg/service/dbms"
	"smart-slowquery/pkg/service/optimize"
	"smart-slowquery/pkg/service/report/email"
	"smart-slowquery/pkg/service/switcher"
	"smart-slowquery/pkg/store/mysql"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.garena.com/shopee/go-shopeelib/gin/middlewares"

	"github.com/astaxie/beego/logs"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	uuid "github.com/satori/go.uuid"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

type PlatformServer struct {
	ConfigFile string
	cfg        *conf.Config
}

func NewPlatformHttpServer(config string) (sv *PlatformServer) {
	return &PlatformServer{
		ConfigFile: config,
		cfg:        &conf.Config{},
	}
}

func (sv *PlatformServer) Start() (err error) {
	var (
		srvApi      *sysHttp.Api
		qs          *service.QueryService
		accSrv      *service.AccessService
		cmdbSrv     *cmdbService.Service
		dbmsSrv     *dbms.DBMetaService
		emailSrv    *email.RankReportService
		rankSrv     *rankService.SlowQueryRankService
		ck          *clickhouse.CKStore
		databaseSrv *cmdbService.DataBasesService
		optimizeSrv *optimize.Service
		mysqlDB     mysql.DB
	)

	// load config
	if err = sv.loadConfig(); err != nil {
		return err
	}

	// init log
	if err = sv.InitLog(); err != nil {
		return err
	}

	log.Info("slow query start ...")
	//init switcher
	switcher.InitOpenSwitcher()

	// optimize-party model init
	{
		thirdAdvisor.InitHeuristicRules()
		thirdCommon.Config.LogLevel = logs.LevelInformational
		thirdCommon.Config.LogOutput = sv.cfg.ServerLog.Path + "/soar.log"
		thirdCommon.LoggerInit()
	}

	// init service id
	if err = sv.InitServiceId(); err != nil {
		return err
	}

	if ck, err = clickhouse.NewCKStore(sv.cfg.CKCli); err != nil {
		return err
	}

	if qs, err = service.NewQueryService(ck); err != nil {
		return err
	}

	// global signal
	var sign os.Signal
	signCh := make(chan os.Signal, 1)
	signal.Notify(signCh, sys.Signs...)

	prometheus := ginPrometheus.NewPrometheus("gin")
	prometheus.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		return c.Request.URL.Path
	}

	engine := gin.Default()
	pprof.Register(engine)
	prometheus.Use(engine)

	engine.Use(
		// get space user
		middlewares.GetUserEmailWithSpaceHost(sv.cfg.SpaceConfig.SpaceHost, sysHttp.APIVersion),
		// http status
		filter.HttpStatus(sysHttp.APIVersion),
		// register with an requestID, in order for log trace
		filter.RequestIDMiddleware(),
		// log
		filter.GinLogger(),
	)

	// create db conn
	if mysqlDB, err = mysql.ConnectToMySQL(sv.cfg.MetaDBConfig); err != nil {
		return fmt.Errorf("init mysql store error:%s \n", err.Error())
	}

	if cmdbSrv, err = cmdbService.NewService(sv.cfg.SpaceConfig); err != nil {
		return fmt.Errorf("init cmdbService error:%s", err.Error())
	}
	//mysql access Service
	if accSrv, err = service.NewAccessService(sv.cfg.MysqlAccessConfig, sv.cfg.MetaDBConfig); err != nil {
		return fmt.Errorf("init remote sql access error: %s", err.Error())
	}

	if dbmsSrv, err = dbms.NewDbMetaService(sv.cfg.CMDBConfig); err != nil {
		return fmt.Errorf("init dbmsService error:%s", err.Error())
	}

	if databaseSrv, err = cmdbService.NewDataBaseService(sv.cfg.CMDBDataBaseConfig); err != nil {
		return fmt.Errorf("init cmdbService.database error:%s", err.Error())
	}

	if emailSrv, err = email.NewRankReportService(sv.cfg.ReportEmailConfig, mysqlDB); err != nil {
		return fmt.Errorf("init emailService error:%s", err.Error())
	}

	if rankSrv, err = rankService.NewSLowQueryRankService(ck, mysqlDB); err != nil {
		return fmt.Errorf("init rankService error:%s", err.Error())
	}
	reportAct := action.NewReportRankAction(rankSrv, dbmsSrv)

	if optimizeSrv, err = optimize.NewService(); err != nil {
		return fmt.Errorf("init optimizeSrv error:%s", err.Error())
	}

	if srvApi, err = sysHttp.NewPlatformApi(engine, qs, accSrv, databaseSrv, dbmsSrv, rankSrv, cmdbSrv, emailSrv, reportAct, optimizeSrv); err != nil {
		return fmt.Errorf("http start api error:%s", err.Error())
	}

	if err = uic.InitUicClient(sv.cfg.SpaceConfig, sv.cfg.UserGroup); err != nil {
		return fmt.Errorf("init uic client error:%s", err.Error())
	}
	//init Prometheus openAPI
	monitor.InitMonitorPrometheusOpenApi(sv.cfg.MonitorPrometheus.OpenApi)

	srvApi.AddPlatformApi(engine)
	router := InitRouter(sv.cfg.ListenPort, sv.cfg.ServerShutdownMaxWaitSeconds)
	router.Run(engine)

	log.Info("slow query start finish")
	// global signal
	sign = <-signCh
	switch sign {
	case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		log.Info("slow query router shutdown ...")
		if err = router.Shutdown(); err != nil {
			log.Error(err)
		}
	}
	return nil
}

func (sv *PlatformServer) loadConfig() (err error) {
	if sv.cfg, err = conf.LoadConfig(sv.ConfigFile); err != nil {
		log.Fatalf("load config error: %v", err)
		return err
	}
	conf.GlobalConfig = *sv.cfg
	return nil
}

func (sv *PlatformServer) InitLog() (err error) {
	// setup log level
	return log.InitLog(sv.cfg.ServerLog)
}

func (sv *PlatformServer) InitServiceId() error {
	id := os.Getenv("")
	log.Infof("get service id by env:%s", id)

	if id == "" {
		id = sv.cfg.ServerName
		log.Infof("get service id by conf:%s", id)
	}

	if id == "" {
		id = uuid.NewV4().String()
		return fmt.Errorf("get service id by random:%s, something will wrong", id)
	}

	conf.ServiceId = id
	return nil
}
