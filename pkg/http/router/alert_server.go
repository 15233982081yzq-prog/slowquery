package router

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"smart-slowquery/internal/util/sys"
	"smart-slowquery/pkg/http/filter"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/oplog"
	"smart-slowquery/pkg/service/alert"
	cmdbService "smart-slowquery/pkg/service/cmdb"
	"smart-slowquery/pkg/service/switcher"
	"smart-slowquery/pkg/service/uic"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/mysql"

	conf "smart-slowquery/conf/alert"
	sysHttp "smart-slowquery/pkg/http"
	ckStore "smart-slowquery/pkg/store/clickhouse"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	uuid "github.com/satori/go.uuid"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

type AlertServer struct {
	ConfigFile string
	cfg        *conf.Config
}

func NewHttpAlertServer(config string) (sv *AlertServer) {
	return &AlertServer{
		ConfigFile: config,
		cfg:        &conf.Config{},
	}
}

func (sv *AlertServer) Start() (err error) {
	var (
		srvApi      *sysHttp.Api
		alertSrv    *alert.Service
		alertOpLog  *oplog.AlertOpLog
		databaseSrv *cmdbService.DataBasesService
		mysqlDB     mysql.DB
		ck          store.CKStore
	)

	// load config
	if err = sv.loadConfig(); err != nil {
		return err
	}

	// init log
	if err = sv.InitLog(); err != nil {
		return err
	}

	log.Info("slow query alert server start ...")
	//init switcher
	switcher.InitOpenSwitcher()

	// init service id
	if err = sv.InitServiceId(); err != nil {
		return err
	}

	if ck, err = ckStore.NewCKStore(sv.cfg.CKCli); err != nil {
		return err
	}

	alertOpLog = oplog.NewAlertOpLog(ck)

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
		filter.AlertApiAuthMiddleware(sv.cfg.SpaceConfig.SpaceHost),
		// http status
		filter.HttpStatus(sysHttp.ALERTAPIVersion),
		// register with an requestID, in order for log trace
		filter.RequestIDMiddleware(),
		// log
		filter.GinLogger(),
	)

	if mysqlDB, err = mysql.ConnectToMySQL(sv.cfg.MetaDBConfig); err != nil {
		return fmt.Errorf("init mysql store error:%s \n", err.Error())
	}

	if databaseSrv, err = cmdbService.NewDataBaseService(sv.cfg.CMDBDataBaseConfig); err != nil {
		return fmt.Errorf("init cmdbService.database error:%s", err.Error())
	}

	if alertSrv, err = alert.NewAlertService(sv.cfg, mysqlDB, ck, alertOpLog); err != nil {
		return fmt.Errorf("init alertSrv error:%s", err.Error())
	}

	if srvApi, err = sysHttp.NewAlertAPIApi(engine, alertSrv, alertOpLog, databaseSrv); err != nil {
		return fmt.Errorf("http start api error:%s", err.Error())
	}

	if err = uic.InitUicClient(sv.cfg.SpaceConfig, sv.cfg.UserGroup); err != nil {
		return fmt.Errorf("init uic client error:%s", err.Error())
	}

	srvApi.AddAlertApi(engine)
	router := InitRouter(sv.cfg.ListenPort, sv.cfg.ServerShutdownMaxWaitSeconds)
	router.Run(engine)

	log.Info("slow query alert server start finish!")
	// global signal
	sign = <-signCh
	switch sign {
	case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		log.Info("slow query alert server router shutdown ...")
		if err = router.Shutdown(); err != nil {
			log.Error(err)
		}
	}
	return nil
}

func (sv *AlertServer) loadConfig() (err error) {
	if sv.cfg, err = conf.LoadConfig(sv.ConfigFile); err != nil {
		log.Fatalf("load config error: %v", err)
		return err
	}
	conf.GlobalConfig = *sv.cfg
	return nil
}

func (sv *AlertServer) InitLog() (err error) {
	// setup log level
	return log.InitLog(sv.cfg.ServerLog)
}

func (sv *AlertServer) InitServiceId() error {
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
