package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smart-slowquery/pkg/action"
	"smart-slowquery/pkg/cron/daily/task"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/dbms"
	"smart-slowquery/pkg/service/report/email"
	"smart-slowquery/pkg/store/mysql"

	conf "smart-slowquery/conf/cronjob"
	cronDaily "smart-slowquery/pkg/cron/daily"
	rankService "smart-slowquery/pkg/service/statistics/daily"
	ckStore "smart-slowquery/pkg/store/clickhouse"
)

const (
	OsEnv              = "ENV"
	ConfigFileTemplate = "/etc/conf/config.%s.toml"
)

var (
	config       = flag.String("config", "", "slow query cronjob config file")
	pprofAddress = flag.String("pprof", ":8098", "")
)

func main() {
	var (
		err        error
		cfg        *conf.Config
		mysqlDB    mysql.DB
		ckRd       *ckStore.CKStore
		dbmsSrv    *dbms.DBMetaService
		rankSrv    *rankService.SlowQueryRankService
		emailSrv   *email.RankReportService
		cronRunner *cronDaily.Runner
	)

	start := time.Now()
	flag.Parse()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, os.Interrupt)

	if cfg, err = conf.LoadConfig(*config); err != nil {
		fmt.Printf("CronJob load config error:%s \n", err.Error())
		return
	}

	//init log
	if err := log.InitLog(cfg.CronLog); err != nil {
		fmt.Printf("CronJob InitLog error:%s \n", err.Error())
		return
	}

	// create db conn
	if mysqlDB, err = mysql.ConnectToMySQL(cfg.MetaDBConfig); err != nil {
		fmt.Printf("CronJob init mysqlDB error:%s \n", err.Error())
		return
	}

	if ckRd, err = ckStore.NewCKStore(cfg.CKCli); err != nil {
		fmt.Printf("CronJob init ckRd error:%s \n", err.Error())
		return
	}

	if dbmsSrv, err = dbms.NewDbMetaService(cfg.CMDBConfig); err != nil {
		fmt.Printf("CronJob init dbMeta error:%s \n", err.Error())
		return
	}

	if rankSrv, err = rankService.NewSLowQueryRankService(ckRd, mysqlDB); err != nil {
		fmt.Printf("CronJob init rankService error:%s \n", err.Error())
		return
	}
	if emailSrv, err = email.NewRankReportService(cfg.ReportEmailConfig, mysqlDB); err != nil {
		fmt.Printf("CronJob init emailService error:%s \n", err.Error())
		return
	}

	if cronRunner, err = cronDaily.NewRunner(initDailyCronTasks(rankSrv, dbmsSrv, emailSrv, cfg.ReportEmailConfig.Env)); err != nil {
		fmt.Printf("CronJob init cronRunner error:%s", err.Error())
		return
	}
	log.Infof("CronRunner start Action")
	cronRunner.Action()

	log.Infof("cronJob run finish time:%v ,cost:%v", time.Now().Local(), time.Since(start))
}

func initDailyCronTasks(rankSrv *rankService.SlowQueryRankService, dbmsService *dbms.DBMetaService, emailService *email.RankReportService, env string) (tasks []task.Task) {
	//tasks = append(tasks, task.NewDBRankTask(rankSrv))
	//tasks = append(tasks, task.NewFingerRankTask(rankSrv))
	tasks = append(tasks, task.NewSendReportEmailTask(action.NewReportRankAction(rankSrv, dbmsService), emailService, env))
	return
}
