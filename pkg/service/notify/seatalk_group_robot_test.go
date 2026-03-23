package notify

import (
	platformConf "smart-slowquery/conf/platform"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var robotId = "h01y_rGcS-Wqm5qDIZDWKA"

func TestMain(m *testing.M) {
	fmt.Println("begin")

	platformConf.GlobalConfig = platformConf.Config{
		SeatalkRobot: nil,
		UserGroup:    nil,
		SpaceConfig: &conf.Space{
			SpaceHost: "https://space.test.shopee.io",
			User:      "db_tools",
			Pass:      "db_tools.bot@shopee.com",
		},
	}
	_ = log.InitLog(&conf.Log{
		Name:     "dml.log",
		Level:    "info",
		Path:     "/tmp/dml_logs",
		Compress: false,
	})
	m.Run()
	fmt.Println("end")
}

func TestMultiPushTextMessage(t *testing.T) {
	robot, err := NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig, robotId)
	assert.Nil(t, err)
	NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig)
	robot.PushTextMessage("unit test 1", false)
	robot.PushTextMessage("unit test 2", false, "jian.bian@shopee.com")
	robot.PushTextMessage("unit test 3", false, "jian.bian@shopee.com", "")
	robot.PushTextMessage("unit test 4", true)
}

func TestPushTextMessage(t *testing.T) {
	robots, err := NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig, robotId, robotId)
	assert.Nil(t, err)
	robots.PushTextMessage("unit test 1, two robotId", false)
}

func TestGetDodMembers(t *testing.T) {
	NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig, robotId)
	r1, _ := NewSeaTalkGroupRobot(true, 0, platformConf.GlobalConfig.SpaceConfig, robotId)
	r1.GetDodMembers()
}

func TestPushTextMessageErr(t *testing.T) {
	r1, err := NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig, "xxqdasdasxasxcaxascacacac")
	assert.Nil(t, err)
	r1.PushTextMessage("unit test 1", false)
}

func TestCTalkDBARobotTest(t *testing.T) {
	r1, _ := NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig, robotId)
	c := &DBANotifyParam{
		IsOk:    false,
		LogicDB: "DDDDD",
		Env:     "EEEE",
		Query: &ExecQuery{
			Sql:          "explain select * from abc",
			ConnectionID: 111,
			Timeout:      true,
			KillHung:     false,
		},
	}
	fmt.Println(GenDBANotifyMarkDown(c))
	r1.PushMarkDownMessage(GenDBANotifyMarkDown(c), false, "jian.bian@shopee.com")
	r1.PushMarkDownMessageAtDod("")
}

func TestCTalkDBADodRobotTest(t *testing.T) {
	//r1, _ := NewSeaTalkGroupRobot(true, 30530, robotId)
	c := &DBANotifyParam{
		IsOk:    false,
		LogicDB: "DDDDD",
		Env:     "EEEE",
		Query: &ExecQuery{
			Sql:          "explain select * from abc",
			ConnectionID: 111,
			Timeout:      true,
			KillHung:     false,
		},
	}
	fmt.Println(GenDBANotifyMarkDown(c))
	c.IsOk = true
	fmt.Println(GenDBANotifyMarkDown(c))
	//r1.PushMarkDownMessageAtDod(GenDBANotifyMarkDown(c))
}

func TestNewSeaTalkGroupRobotByNoRobotIds(t *testing.T) {
	_, err := NewSeaTalkGroupRobot(false, 0, nil, []string{}...)
	assert.Error(t, err)
}

func TestNewSeaTalkGroupRobotByNilConfig(t *testing.T) {
	_, err := NewSeaTalkGroupRobot(false, 0, nil, robotId)
	assert.Error(t, err)
}

func TestNewSeaTalkGroupRobotSucc(t *testing.T) {
	r1, err := NewSeaTalkGroupRobot(false, 0, platformConf.GlobalConfig.SpaceConfig, robotId)
	assert.NoError(t, err)
	assert.NotNil(t, r1)
}
