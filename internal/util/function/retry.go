package function

import (
	"smart-slowquery/pkg/log"

	"time"
)

const (
	defRetry = 2

	// MaxRetrySleepGap 最大重试时间间隔
	MaxRetrySleepGap = time.Second
)

type RetryFunc func() error

// Retry 开始重试
func Retry(funcName string, rf RetryFunc, times int) (err error) {
	if times < defRetry {
		times = defRetry
	}
	for i := 0; i <= times; i++ {
		if err = rf(); err != nil {
			log.Errorf("func:%s ,retry i:%d error:%s \n", funcName, i, err.Error())
			time.Sleep(time.Duration(i*i) * MaxRetrySleepGap)
			continue
		}
		break
	}
	return nil
}
