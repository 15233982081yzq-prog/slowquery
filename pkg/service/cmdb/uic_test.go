package cmdb

import (
	"context"
	"fmt"
	"testing"

	"smart-slowquery/conf"
	platformConf "smart-slowquery/conf/platform"
	"smart-slowquery/pkg/log"

	"git.garena.com/shopee/platform/space-sdk/uic"
	"github.com/stretchr/testify/assert"
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

	platformConf.GlobalConfig = platformConf.Config{
		SeatalkRobot: nil,
		UserGroup:    nil,
		SpaceConfig: &conf.Space{
			SpaceHost: "https://space.test.shopee.io",
			User:      "db_tools",
			Pass:      "db_tools.bot@shopee.com",
		},
	}

	if ret := t.Run(); ret != 0 {
		log.Error(fmt.Errorf("ret is %d", ret))
		return
	}
}

func TestGetUserGroup(t *testing.T) {
	c, err := NewUicClient(platformConf.GlobalConfig.SpaceConfig)
	assert.NoError(t, err)

	r := uic.GetGroupMemberReq{
		GroupIDs: []uint64{50917},
	}
	res, e := c.GetGroupMembers(context.Background(), r)
	assert.NoError(t, e)
	assert.NotNil(t, res)
	fmt.Println(res)
}
