package conf

import (
	"github.com/jinzhu/configor"
	"smart-slowquery/conf"
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
	DBDetailUrl                  string `toml:"db_detail_url"`
	OpenApiToken                 string `toml:"open_api_token"`
	ServerShutdownMaxWaitSeconds int    `toml:"server_shutdown_max_wait_seconds"`

	// ------------ DB_CK -----------------
	CKCli *conf.CKCli `toml:"clickhouse_client_config"`
	// ------------- Log ------------------
	ServerLog *conf.Log `toml:"server_log_config"`

	SpaceConfig *conf.Space `toml:"space_config"`
	//--------------Http_Rate_Limit --------
	HttpRateLimit *conf.RateLimitConfig `toml:"http_rate_limit"`
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
