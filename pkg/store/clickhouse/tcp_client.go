package clickhouse

import (
	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"

	"fmt"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func NewTcpClient(config *conf.CKCli) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("clickhouse config is empty")
	}
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=5s&max_execution_time=30", config.User, config.Password, config.Host, config.Port, config.DbName)
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
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
	log.Infof("clickhouse NewTcpClient success,address:%s", fmt.Sprintf("%s:%d", config.Host, config.Port))
	return cli, nil
}
