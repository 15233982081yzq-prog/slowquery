package log

import (
	"smart-slowquery/conf"

	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	zaplogfmt "github.com/sykesm/zap-logfmt"
	uzap "go.uber.org/zap"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	czap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type logger struct {
	lgr logr.Logger

	level int // debug, info, warn, error => 5, 4, 3, 2, 1

	jobName, jobUuid string
}

const (
	LevelAll   = 10
	LevelDebug = 7
	LevelInfo  = 6
	LevelWarn  = 5
	LevelError = 4
	LevelFatal = 3

	logPath     = "/tmp/log"
	maxFileSize = 100
	maxAge      = 7
	maxBackups  = 7
)

var (
	// Log 与所有调用方保持一致
	log *logger
)

func logName() string {
	bin := os.Args[0]
	if strings.Contains(bin, "-") {
		// 去掉 - 字符
		return bin[strings.LastIndex(bin, "-")+1:]
	}

	return bin + "-" + log.jobUuid
}

func InitLog(conf *conf.Log) error {
	//path，level,
	if conf == nil {
		return fmt.Errorf("log config is nil")
	}
	validConfig(conf)

	log = &logger{}
	log.lgr = logf.Log.WithName(conf.Name)
	log.level = getLvl(conf.Level)
	initZapLogger(conf)
	return nil
}

func validConfig(conf *conf.Log) {
	if conf.Name == "" {
		conf.Name = logName()
	}
	if conf.Path == "" {
		conf.Path = logPath
	}
	if conf.MaxFileSize == 0 {
		conf.MaxFileSize = maxFileSize
	}
	if conf.MaxBackups == 0 {
		conf.MaxBackups = maxBackups
	}
	if conf.MaxAge == 0 {
		conf.MaxAge = maxAge
	}
}

func (l *logger) SetJobName(jobName string) {
	l.jobName = jobName
}

func (l *logger) SetJobUuid(jobUuid string) {
	l.jobUuid = jobUuid
}

// to be compatible to log platform
// original msg format:   ts=xx level=xx logger=xx msg=xx
// after msg format:      ts=xx level=xx logger=xx msg={jobName:xx, jobUuid:xx, msg:xx}
// original error format: ts=xx level=xx logger=xx msg= error=xx
// after error format:    ts=xx level=xx logger=xx msg={jobName:xx, jobUuid:xx, msg:error} error=xx
func (l *logger) logMessageDecorator(msg string, logLevel int) string {
	// dmsg : decoratored msg
	var dmsg string
	dmsg = "{msg:"
	// set error's msg = error
	if logLevel == LevelError {
		dmsg += "error"
	} else {
		dmsg += msg
	}
	dmsg += "}"
	return dmsg
}

func getLvl(level string) int {
	switch level {
	case "all":
		return LevelAll
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelAll
	}
}

func IsDebugEnabled() bool {
	return log.level >= LevelDebug
}

func Errorf(format string, vals ...interface{}) {
	if log.level < LevelError {
		return
	}
	log.lgr.Error(fmt.Errorf(format, vals...), log.logMessageDecorator("", LevelError))
}

func Error(err error) {
	if log.level < LevelError {
		return
	}
	log.lgr.Error(err, log.logMessageDecorator("", LevelError))
}

func Debugf(format string, vals ...interface{}) {
	if log.level < LevelDebug {
		return
	}
	msg := log.logMessageDecorator(fmt.Sprintf(format, vals...), LevelDebug)
	log.lgr.Info(msg)
}

func Debug(msg string) {
	if log.level < LevelDebug {
		return
	}
	log.lgr.Info(log.logMessageDecorator(msg, LevelDebug))
}

func Warningf(format string, vals ...interface{}) {
	if log.level < LevelWarn {
		return
	}
	msg := log.logMessageDecorator(fmt.Sprintf(format, vals...), LevelWarn)
	log.lgr.Info(msg)
}

func Warning(msg string) {
	if log.level < LevelWarn {
		return
	}
	log.lgr.Info(log.logMessageDecorator(msg, LevelWarn))
}

func Infof(format string, vals ...interface{}) {
	if log.level < LevelInfo {
		return
	}
	msg := log.logMessageDecorator(fmt.Sprintf(format, vals...), LevelInfo)
	log.lgr.Info(msg)
}

func Info(msg string) {
	if log.level < LevelInfo {
		return
	}
	log.lgr.Info(log.logMessageDecorator(msg, LevelInfo))
}

func Fatalf(format string, vals ...interface{}) {
	if log.level < LevelFatal {
		return
	}
	msg := log.logMessageDecorator(fmt.Sprintf(format, vals...), LevelFatal)
	log.lgr.Info(msg)
	os.Exit(1)
}

// initZapLogger 设置 zap 日志
func initZapLogger(conf *conf.Log) {
	encoder := getEncoder()
	writerSync := getLogWriter(conf)

	logger := czap.New(czap.UseDevMode(true),
		czap.WriteTo(writerSync),
		czap.Encoder(encoder))
	logf.SetLogger(logger)
}

func getLogWriter(conf *conf.Log) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path.Join(conf.Path, fmt.Sprintf("%s.log", conf.Name)),
		MaxSize:    conf.MaxFileSize,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,
		Compress:   conf.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func getEncoder() zapcore.Encoder {
	cl := uzap.NewProductionEncoderConfig()
	cl.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.Local().Format(time.RFC3339Nano))
	}
	return zaplogfmt.NewEncoder(cl)
}
