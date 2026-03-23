package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	conf "smart-slowquery/conf/analyzer"
	mysqlAnalyzer "smart-slowquery/pkg/analyzer"
	"smart-slowquery/pkg/service/dbms"
	"smart-slowquery/pkg/store/mysql"

	"smart-slowquery/pkg/kafka"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/clickhouse"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cfgPath      = flag.String("config", "", "")
	pprofAddress = flag.String("pprof", ":8098", "")
	pattern      = `/\*+\s*%{WORD:directive}\s+%{WORD:trace_info}\s*\*/`
)

func main() {
	var (
		ckCli     *clickhouse.Client
		kfkCli    *kafka.Client
		service   *mysqlAnalyzer.Service
		dbmsSrv   *dbms.DBMetaService
		cfg       *conf.Config
		monitorDB mysql.DB
		err       error
	)

	start := time.Now()
	flag.Parse()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, os.Interrupt)
	//server error chan
	servc := make(chan error, 1)

	if cfg, err = conf.LoadConfig(*cfgPath); err != nil {
		fmt.Printf("load config error:%s \n", err.Error())
		return
	}

	//init log
	if err = log.InitLog(cfg.Log); err != nil {
		fmt.Printf("InitLog error:%s \n", err.Error())
		return
	}
	//创建一个连接到已运行的ClickHouse服务的客户端，而不是启动ClickHouse服务本身。ClickHouse服务应该已经在外部部署并运行
	if ckCli, err = clickhouse.NewClient(cfg.CKCli); err != nil {
		fmt.Printf("ck ckCli error:%s \n", err.Error())
		return
	}

	if monitorDB, err = mysql.ConnectToMySQL(cfg.MonitorDBConfig); err != nil {
		fmt.Printf("init mysql store error:%s \n", err.Error())
		return
	}

	if dbmsSrv, err = dbms.NewDbMetaService(cfg.CMDBConfig); err != nil {
		fmt.Printf("init dbmsService error:%s \n", err.Error())
		return
	}

	if service, err = mysqlAnalyzer.NewService(cfg, true, ckCli, monitorDB, dbmsSrv); err != nil {
		fmt.Printf("mysqlAnalyzer.NewService error:%s \n", err.Error())
		return
	}
	//创建一个连接到已运行的Kafka服务的客户端
	if kfkCli, err = kafka.NewClient(cfg.Kafka, service); err != nil {
		fmt.Printf("kafka.NewClient error:%s \n", err.Error())
		return
	}

	if err = kfkCli.Consume(); err != nil {
		fmt.Printf("kfkCli.Consume error:%s \n", err.Error())
		return
	}

	//prometheus metrics http
	http.Handle("/metrics", promhttp.Handler())
	//pprof
	pprof := &http.Server{Addr: *pprofAddress, Handler: nil}
	go func() {
		servc <- pprof.ListenAndServe()
	}()
	log.Info("pprof init ok!")

	log.Infof("analyzer server start finish time:%v ,cost:%v", time.Now().Local(), time.Since(start))

	select {
	case <-signals:
		goto EXIT
	case <-servc:
		goto EXIT
	}

EXIT:
	//close service
	kfkCli.Close()
	service.Close()
	ckCli.Close()
	log.Infof("Interrupt signal received closed time:%v ,Run for:%v", time.Now().Local(), time.Since(start))
}
