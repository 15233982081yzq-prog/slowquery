package optimize

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	mysqld "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/mysql"
)

type TestTable struct {
	ID         int       `gorm:"column:id"`
	UserName   string    `gorm:"column:user_name"`
	DBName     string    `gorm:"column:db_name"`
	DBEnv      string    `gorm:"column:db_env"`
	Age        int       `gorm:"column:age"`
	Info       string    `gorm:"column:info;type:text"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoCreateTime"`
}

type TestTableWithIndex struct {
	ID         int       `gorm:"column:id"`
	UserName   string    `gorm:"column:user_name;index:idx_test"`
	DBName     string    `gorm:"column:db_name"`
	DBEnv      string    `gorm:"column:db_env"`
	Age        int       `gorm:"column:age;index:idx_test"`
	Info       string    `gorm:"column:info;type:text"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoCreateTime"`
}

type testTable1K TestTable
type testTable1KWithIndex TestTableWithIndex

func (t1 testTable1K) TableName() string {
	return "table_1k"
}

type testTable1W TestTable

func (t2 testTable1W) TableName() string {
	return "table_1w"
}

func (t1W testTable1KWithIndex) TableName() string {
	return "table_1w_with_index"
}

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
			"MYSQL_DATABASE=" + dbName,
		},
		Config: &conf.MetaDBConfig{
			Username:       "root",
			Password:       "root",
			Host:           "localhost",
			DBName:         "test",
			LogMode:        true,
			ConnectTimeout: 20,
			MaxIdleConns:   10,
			MaxOpenConns:   50,
			ErrRetry:       3,
		},
	}
	dbName         = "test"
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
		if err = db.AutoMigrate(&testTable1K{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&testTable1W{}); err != nil {
			return err
		}
		if err = db.AutoMigrate(&testTable1KWithIndex{}); err != nil {
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
