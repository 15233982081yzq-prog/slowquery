package clickhouse

import (
	"fmt"
	"time"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"

	"github.com/ClickHouse/clickhouse-go/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	gormCK "gorm.io/driver/clickhouse"
)

type Client struct {
	db  *gorm.DB
	cfg *conf.CKCli
}

func NewClient(config *conf.CKCli) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("clickhouse config is empty")
	}

	opt := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.DbName,
			Username: config.User,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 120,
		},
		DialTimeout: 60 * time.Second,
		Protocol:    clickhouse.HTTP,
		//MaxOpenConns: 20,
		//MaxIdleConns: 20,
		ReadTimeout: 300 * time.Second,
	}

	if config.Compression {
		opt.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		}
	}

	conn := clickhouse.OpenDB(opt)
	db, err := gorm.Open(gormCK.New(gormCK.Config{
		Conn: conn, // initialize with existing database conn
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Errorf("Failed to connect to ClickHouse:%s", err.Error())
		return nil, err
	}

	ckDB, err := db.DB()
	if err != nil {
		log.Errorf("gorm ClickHouse db.DB() error:%s", err.Error())
		return nil, err
	}

	ckDB.SetMaxIdleConns(config.MaxIdleConn)
	ckDB.SetMaxOpenConns(config.MaxOpenConn)
	ckDB.SetConnMaxLifetime(config.MaxLifeTime.Duration)

	cli := &Client{
		db:  db,
		cfg: config,
	}
	log.Infof("clickhouse new http ckCli success,address:%s", fmt.Sprintf("%s:%d", config.Host, config.Port))
	return cli, nil
}

func (cli *Client) Close() {
	db, _ := cli.db.DB()
	db.Close()
}
