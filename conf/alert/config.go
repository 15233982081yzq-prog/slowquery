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
	ListenPort                   int      `toml:"listen_port"`
	GinMode                      string   `toml:"gin_mode"`
	DBDetailUrl                  string   `toml:"db_detail_url"`
	CallBackToken                string   `toml:"callback_token"`
	CallBackUrl                  string   `toml:"call_back_url"`
	ServerShutdownMaxWaitSeconds int      `toml:"server_shutdown_max_wait_seconds"`
	Admins                       []string `toml:"admins"`

	// ------------ DB_CK -----------------
	CKCli *conf.CKCli `toml:"clickhouse_client_config"`
	// ------------- Log ------------------
	ServerLog *conf.Log `toml:"server_log_config"`
	// ------------- Space ------------------
	SpaceConfig *conf.Space `toml:"space_config"`
	// ------------- DB ------------------
	MetaDBConfig *conf.MetaDBConfig `toml:"db_config"`
	// ------------- User Group ------------------
	UserGroup *conf.UserGroupConfig `toml:"user_group_config"`
	// ------------- Monitor ----------------
	MonitorClientConfig *conf.MonitorClientConfig `toml:"monitor_client_config"`
	// ------------- Dod ----------------
	DodConfig *conf.DodConfig `toml:"dod_config"`
	// ------------- cmdb db ----------------
	CMDBDataBaseConfig *conf.CMDBDataBase `toml:"cmdb_database_config"`
	//--------------Alert template --------
	AlertTemplates map[string]*conf.AlertTemplate `toml:"alert_templates"`
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
