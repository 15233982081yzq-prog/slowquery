package function

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	"testing"
)

func TestMain(t *testing.M) {
	// Init log, tem config
	_ = log.InitLog(&conf.Log{
		Name:        "test",
		Level:       "info",
		Path:        "./tmp/log",
		MaxFileSize: 500,
		MaxBackups:  10,
		MaxAge:      10,
		Compress:    true,
	})

	if ret := t.Run(); ret != 0 {
		log.Error(fmt.Errorf("ret is %d", ret))
		return
	}
}

func TestRetryNoError(t *testing.T) {
	err := Retry("TestRetryNoError", func() error {
		fmt.Printf("test retry no error\n")
		return nil
	}, 0)
	assert.NoError(t, err)
}

func TestRetryError(t *testing.T) {
	err := Retry("TestRetryError", func() error {
		fmt.Printf("test retry error\n")
		return fmt.Errorf("return error")
	}, 1)
	assert.Equal(t, err, nil)
}
