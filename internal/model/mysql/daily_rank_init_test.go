package mysql

import (
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	mysqld "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/mysql"
	"strconv"
	"testing"
	"time"
)

type TestDBSetting struct {
	Config       *conf.MetaDBConfig
	ImageName    string
	ImageVersion string
	ENV          []string
}

var (
	dbSetting = &TestDBSetting{
		ImageName:    "percona/percona-server",
		ImageVersion: "5.7",
		ENV: []string{
			"MYSQL_ROOT_PASSWORD=" + "root",
			"MYSQL_DATABASE=" + "shopee_exec_meta_db",
		},
		Config: &conf.MetaDBConfig{
			Username:       "root",
			Password:       "root",
			Host:           "localhost",
			DBName:         "shopee_exec_meta_db",
			LogMode:        true,
			ConnectTimeout: 20,
			MaxIdleConns:   10,
			MaxOpenConns:   50,
			ErrRetry:       3,
		},
	}
	port     int32 = 3306
	tearDown func()
	dbClient mysql.DB
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

// if you using colima to run docker in your macbook
// Pleases exec: sudo ln -sf $HOME/.colima/default/docker.sock /var/run/docker.sock
// ----
// if you locally exist mysql, please use port 3306
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

	containerPort := resource.GetPort(fmt.Sprintf("%d/tcp", port))
	if portNum, err := strconv.Atoi(containerPort); err != nil || portNum != 0 {
		port = int32(portNum)
	}

	if err := pool.Retry(func() error {
		dbSetting.Config.Port = port
		dsn := fmt.Sprintf(mysql.MySQLDSNFmt, dbSetting.Config.Username, dbSetting.Config.Password,
			dbSetting.Config.Host, dbSetting.Config.Port, dbSetting.Config.DBName, dbSetting.Config.ConnectTimeout)
		db, err := gorm.Open(mysqld.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
		if err := db.AutoMigrate(&DBSlowQueryDailyRank{}); err != nil {
			return err
		}
		if err := db.AutoMigrate(&FingerSlowQueryDailyRank{}); err != nil {
			return err
		}
		if err := db.AutoMigrate(&DailyReportLog{}); err != nil {
			return err
		}
		if err := db.AutoMigrate(&DailyNewFingerReportLog{}); err != nil {
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
	dbSetting.Config.Port = port

	// init tem clickhouse OLAP database
	db, err := initDatabase(dbSetting)
	if err != nil {
		return nil, err
	}
	dbClient = db

	return func() { _ = db.Close() }, nil
}

func initDatabase(dbSetting *TestDBSetting) (mysql.DB, error) {
	return mysql.ConnectToMySQL(dbSetting.Config)
}
