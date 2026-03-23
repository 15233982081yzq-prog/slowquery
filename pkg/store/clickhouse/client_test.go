package clickhouse

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"
)

const (
	// image info:https://hub.docker.com/r/clickhouse/clickhouse-server/
	clickhouseImage    = "clickhouse/clickhouse-server"
	clickhouseVersion  = "23.9.5"
	clickhouseDB       = "szinfra_clouddba_slow_query_test"
	clickhouseUser     = "default"
	clickhousePassword = ""
)

type TestDBSetting struct {
	Config       *conf.CKCli
	ImageName    string
	ImageVersion string
	ENV          []string
}

var (
	ckReadClient *CKStore
	ckClient     *Client
	dbSetting    = &TestDBSetting{
		ImageName:    clickhouseImage,
		ImageVersion: clickhouseVersion,
		ENV: []string{
			"CLICKHOUSE_USER=" + clickhouseUser,
			"CLICKHOUSE_PASSWORD=" + clickhousePassword,
			"CLICKHOUSE_DB=" + clickhouseDB},
		Config: &conf.CKCli{
			Driver:      "clickhouse",
			Host:        "localhost",
			User:        clickhouseUser,
			Password:    clickhousePassword,
			DbName:      clickhouseDB,
			MaxOpenConn: 30,
			MaxIdleConn: 10,
		},
	}
	clickhouseHTTPPort int32 = 8123
	clickhouseTCPPort  int32 = 9000
	tearDown           func()
)

func TestMain(t *testing.M) {
	defer func() {
		if tearDown != nil {
			tearDown()
		}
	}()

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

	cleanImageFunc, pullImageErr := initTestImage(dbSetting)

	cleanDataFunc, err := initTestDataSource(dbSetting)
	if err != nil {
		panic(err)
		return
	}

	tearDown = func() {
		// if this test in container, need`t exec cleanDataFunc()
		cleanDataFunc()
		cleanImageFunc(pullImageErr == nil)
	}

	if ret := t.Run(); ret != 0 {
		log.Error(fmt.Errorf("ret is %d", ret))
		return
	}
}

// if you use colima to run docker in your macbook
// Pleases exec: sudo ln -sf $HOME/.colima/default/docker.sock /var/run/docker.sock
// ----
// if you locally exist clickhouse, please use port 8123 and 9000
// ----
// if you using podman, please use docker drive，Just execute it directly
func initTestImage(dbSetting *TestDBSetting) (cleanFunc func(bool), err error) {
	pool, err := dockertest.NewPool("")
	pool.MaxWait = time.Minute * 5
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: dbSetting.ImageName,
		Tag:        dbSetting.ImageVersion,
		Env:        dbSetting.ENV,
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("could not pull resource: %s", err)
	}

	httpPort := resource.GetPort(fmt.Sprintf("%d/tcp", clickhouseHTTPPort))
	if portNum, err := strconv.Atoi(httpPort); err != nil || portNum != 0 {
		clickhouseHTTPPort = int32(portNum)
	}

	tcpPort := resource.GetPort(fmt.Sprintf("%d/tcp", clickhouseTCPPort))
	if portNum, err := strconv.Atoi(tcpPort); err != nil || portNum != 0 {
		clickhouseTCPPort = int32(portNum)
	}

	if err := pool.Retry(func() error {
		dbSetting.Config.Port = clickhouseHTTPPort
		_, err := NewClient(dbSetting.Config)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not connect to database: %s", err)
	}
	return func(do bool) {
		if do {
			_ = pool.Purge(resource)
		}
	}, nil

}

func initTestDataSource(dbSetting *TestDBSetting) (cleanFunc func(), err error) {
	// init port info
	dbSetting.Config.Port = clickhouseHTTPPort

	// init tem clickhouse OLAP database
	ckReader, err := initDatabaseRead(dbSetting)
	if err != nil {
		return nil, err
	}
	ckReadClient = ckReader

	ckWrite, err := initDatabaseReadWrite(dbSetting)
	if err != nil {
		return nil, err
	}
	ckClient = ckWrite

	if err = initDatabaseTable(ckReader); err != nil {
		return nil, err
	}

	fakerData, err := initDatabaseData(ckClient)
	if err != nil {
		return nil, err
	}

	return func() { _ = cleanDatabaseData(ckReader, fakerData) }, nil
}

func initDatabaseRead(dbSetting *TestDBSetting) (*CKStore, error) {
	return NewCKStore(dbSetting.Config)
}

func initDatabaseTable(ck *CKStore) error {
	if !ck.client.db.Migrator().HasTable(&request.SlowQueryLog{}) {
		ck.client.db.Migrator().CreateTable(&request.SlowQueryLog{})
	}
	if !ck.client.db.Migrator().HasTable(&request.AlertMessage{}) {
		ck.client.db.Migrator().CreateTable(&request.AlertMessage{})
	}
	if !ck.client.db.Migrator().HasTable(&request.AlertOperatorLog{}) {
		ck.client.db.Migrator().CreateTable(&request.AlertOperatorLog{})
	}
	if !ck.client.db.Migrator().HasTable(&request.SlowQueryLog{}) {
		ck.client.db.Migrator().CreateTable(&request.SlowQueryLog{})
	}
	if !ck.client.db.Migrator().HasTable(&request.AlertMute{}) {
		ck.client.db.Migrator().CreateTable(&request.AlertMute{})
	}
	return nil
}

func TestCleanTable(t *testing.T) {
	ckClient.db.Where("1=1").Delete(&response.SlowQuery{})
}
