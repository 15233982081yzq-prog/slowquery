package conf

type RemoteMysqlAccessConfig struct {
	IsTest       bool   `toml:"is_test"`
	User         string `toml:"user"`
	TestPassword string `toml:"test_password"`
	Key          string `toml:"key"`
}

type MetaDBConfig struct {
	Username       string `toml:"username"`
	Password       string `toml:"password"`
	Host           string `toml:"host"`
	Port           int32  `toml:"port"`
	DBName         string `toml:"db_name"`
	LogMode        bool   `toml:"log_mode"`
	ConnectTimeout int    `toml:"connect_timeout"`
	MaxIdleConns   int    `toml:"max_idle_conns"`
	MaxOpenConns   int    `toml:"max_open_conns"`
	ErrRetry       int    `toml:"error_retry"`
}
