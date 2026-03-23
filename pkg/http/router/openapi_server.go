package router

import (
	conf "smart-slowquery/conf/openapi"
	sysHttp "smart-slowquery/pkg/http"
	ckStore "smart-slowquery/pkg/store/clickhouse"

	"smart-slowquery/internal/util/sys"
	"smart-slowquery/pkg/http/filter"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service"
	"smart-slowquery/pkg/service/switcher"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"

	uuid "github.com/satori/go.uuid"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

type OpenApiServer struct {
	ConfigFile string
	cfg        *conf.Config
}

func NewHttpOpenApiServer(config string) (sv *OpenApiServer) {
	return &OpenApiServer{
		ConfigFile: config,
		cfg:        &conf.Config{},
	}
}

func (sv *OpenApiServer) Start() (err error) {
	var (
		srvApi *sysHttp.Api
		qs     *service.QueryService
		ck     *ckStore.CKStore
	)

	// load config
	if err = sv.loadConfig(); err != nil {
		return err
	}

	// init log
	if err = sv.InitLog(); err != nil {
		return err
	}

	log.Info("slow query open_api start ...")
	//init switcher
	switcher.InitOpenSwitcher()

	// init service id
	if err = sv.InitServiceId(); err != nil {
		return err
	}

	if ck, err = ckStore.NewCKStore(sv.cfg.CKCli); err != nil {
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
		filter.OpenApiAuthMiddleware(sysHttp.OPENAPIVersion),
		// http status
		filter.HttpStatus(sysHttp.OPENAPIVersion),
		// register with an requestID, in order for log trace
		filter.RequestIDMiddleware(),
		// log
		filter.GinLogger(),
	)

	limiter := ratelimit.NewBucketWithRate(sv.cfg.HttpRateLimit.Rate, sv.cfg.HttpRateLimit.Capacity)
	log.Infof("init rate limiter rate:%f,capacity:%d", sv.cfg.HttpRateLimit.Rate, sv.cfg.HttpRateLimit.Capacity)

	if srvApi, err = sysHttp.NewOpenAPIApi(engine, qs); err != nil {
		return fmt.Errorf("http start api error:%s", err.Error())
	}
	srvApi.AddOpenApi(engine, limiter)
	router := InitRouter(sv.cfg.ListenPort, sv.cfg.ServerShutdownMaxWaitSeconds)
	router.Run(engine)

	log.Info("slow query open_api start finish")
	// global signal
	sign = <-signCh
	switch sign {
	case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		log.Info("slow query open_api router shutdown ...")
		if err = router.Shutdown(); err != nil {
			log.Error(err)
		}
	}
	return nil
}

func (sv *OpenApiServer) loadConfig() (err error) {
	if sv.cfg, err = conf.LoadConfig(sv.ConfigFile); err != nil {
		log.Fatalf("load config error: %v", err)
		return err
	}
	conf.GlobalConfig = *sv.cfg
	return nil
}

func (sv *OpenApiServer) InitLog() (err error) {
	// setup log level
	return log.InitLog(sv.cfg.ServerLog)
}

func (sv *OpenApiServer) InitServiceId() error {
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
