package function

import (
	"smart-slowquery/pkg/log"

	"time"
)

type LoopFunc func() error

func Loop(funcName string, lf LoopFunc, sleep time.Duration) {
	num := 1
	for {
		start := time.Now()
		err := lf()
		log.Infof("func:%s ,loop:%d ,cost:%s ,error:%v", funcName, num, time.Since(start), err)
		num++
		time.Sleep(sleep)
	}
}
