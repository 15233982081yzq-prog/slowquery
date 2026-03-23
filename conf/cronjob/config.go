package conf

import (
	"smart-slowquery/conf"

	"github.com/jinzhu/configor"
)

var GlobalConfig Config
var ServiceId string

const RunEnv = "env"

type Config struct {
	// ------------ Server meta ------------
	ServerName string `toml:"server_name"`
	ServerEnv  string `toml:"server_env"`

	// ------------ HTTP ---------------
	ListenPort                   int    `toml:"listen_port"`
	GinMode                      string `toml:"gin_mode"`
	ServerShutdownMaxWaitSeconds int    `toml:"server_shutdown_max_wait_seconds"`

	// ------------ DB_CK -----------------
	CKCli *conf.CKCli `toml:"clickhouse_client_config"`
	// ------------- Log ------------------
	CronLog *conf.Log `toml:"cron_log_config"`

	SpaceConfig *conf.Space `toml:"space_config"`

	CMDBConfig *conf.CMDBConfig `toml:"cmdb_config"`

	CMDBDataBaseConfig *conf.CMDBDataBase `toml:"cmdb_database_config"`

	MetaDBConfig *conf.MetaDBConfig `toml:"db_config"`

	ReportEmailConfig *conf.ReportEmailConfig `toml:"report_email"`
}

func LoadConfig(configFile string) (*Config, error) {
	config := &Config{}
	if err := configor.Load(config, configFile); err != nil {
		return nil, err
	}
	return config, nil
}

type Server struct {
	Name         string `toml:"name"`
	Path         string `toml:"path"`
	HttpMethod   string `toml:"http_method"`
	HttpProtocol string `toml:"http_protocol"`
}
