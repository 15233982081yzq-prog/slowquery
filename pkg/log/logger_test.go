package log

import (
	conf2 "smart-slowquery/conf"
	"smart-slowquery/conf/platform"

	"fmt"
	"strconv"
	"testing"
)

var (
	dir = "./test/log-config.test.toml"
	cfg *conf.Config
)

func TestLogAll(t *testing.T) {
	InitLog(cfg.ServerLog)
	for i := 0; i < 100; i++ {
		Debug("debug log " + strconv.Itoa(i))
		Debugf("生成debug测试日志 %d", i)
		Info("info log " + strconv.Itoa(i))
		Infof("生成info测试日志 %d", i)
		Warning("warning log " + strconv.Itoa(i))
		Warningf("生成warning测试日志 %d", i)
		Error(fmt.Errorf("error log %d", i))
		Errorf("生成error测试日志 %d", i)
	}
	initZapLogger(cfg.ServerLog)
}

func TestLogAll1(t *testing.T) {
	InitLog(&conf2.Log{
		Name:        "tmp",
		Level:       "0",
		Path:        "/",
		MaxFileSize: 0,
		MaxBackups:  0,
		MaxAge:      0,
		Compress:    false,
	})
	for i := 0; i < 100; i++ {
		Debug("debug log " + strconv.Itoa(i))
		Debugf("生成debug测试日志 %d", i)
		Info("info log " + strconv.Itoa(i))
		Infof("生成info测试日志 %d", i)
		Warning("warning log " + strconv.Itoa(i))
		Warningf("生成warning测试日志 %d", i)
		Error(fmt.Errorf("error log %d", i))
		Errorf("生成error测试日志 %d", i)
	}
}

func TestLogName(t *testing.T) {
	InitLog(cfg.ServerLog)
	nm := logName()
	fmt.Printf("logName:%s \n", nm)
}

func TestSetJobName(t *testing.T) {
	InitLog(cfg.ServerLog)
	log.SetJobName("job_name")
}

func TestJobUUid(t *testing.T) {
	InitLog(cfg.ServerLog)
	log.SetJobUuid("ssss")
	log.logMessageDecorator("ssss", 1)
}

func TestGetLogLvl(t *testing.T) {
	array := []string{"all", "debug", "info", "warn", "error", "fatal", "other"}
	for _, one := range array {
		getLvl(one)
	}
}

func TestIsDebugEnabled(t *testing.T) {
	IsDebugEnabled()
}

func init() {
	cfg, _ = conf.LoadConfig(dir)
}
