package conf

import (
	"smart-slowquery/conf"

	"github.com/jinzhu/configor"
)

var slowPattern = map[string]string{
	"GREEDYMULTILINE": `(.|\n)*`,
	"METRICSPACE":     `([ #\n]*)`,
	"EXPLAIN":         `(# explain:.*\n|#\\s*\n)*`,
	"SLOW": "^# User@Host: %{USER:defaultUser}(\\[%{USER:user}\\])?%{METRICSPACE}@%{METRICSPACE}%{HOSTNAME:clientHost}?%{METRICSPACE}\\[%{IP:clientIP}?\\]%{METRICSPACE}(Id:%{SPACE}%{NUMBER:connectionId:int}%{METRICSPACE})?(Thread_id:%{SPACE}%{NUMBER:connectionId:int}%{METRICSPACE})?" +
		"(Schema:%{SPACE}%{WORD:currentDB}?%{METRICSPACE})?(Last_errno: %{NUMBER:lastErrno:int}%{METRICSPACE})?(Killed: %{NUMBER:killed:int}%{METRICSPACE})?(QC_hit: %{WORD:queryCacheHit}%{METRICSPACE})?" +
		"(Query_time: %{NUMBER:queryTime:float}%{METRICSPACE})?(Lock_time: %{NUMBER:lockTime:float}%{METRICSPACE})?(Rows_sent: %{NUMBER:numRows:int}%{METRICSPACE})?(Rows_examined: %{NUMBER:examinedRows:int}%{METRICSPACE})?(Rows_affected: %{NUMBER:affectedRows:int}%{METRICSPACE})?" +
		"(Thread_id: %{NUMBER:connectionId:int}%{METRICSPACE})?(Errno: %{NUMBER:lastErrno:int}%{METRICSPACE})?(Killed: %{NUMBER:killed:int}%{METRICSPACE})?(Bytes_sent: %{NUMBER:bytesSent:int}%{METRICSPACE})?" +
		"%{EXPLAIN:explain}?(use %{WORD:currentDB};\n)?SET timestamp=%{NUMBER:timestamp:int};\n" +
		"%{GREEDYMULTILINE:query}",
}

var GlobalConfig Config

type Config struct {
	Basic             *Basic                        `toml:"basic"`
	CKCli             *conf.CKCli                   `toml:"clickhouse_client_config"`
	CKFlush           *conf.CKFlusher               `toml:"clickhouse_flusher_config"`
	Kafka             *conf.Kafka                   `toml:"kafka_config"`
	Filter            *conf.FingerFilter            `toml:"filter_config"`
	Fresh             *conf.FreshFingerFilter       `toml:"fresh_filter_config"`
	Analyzer          *Analyzer                     `toml:"analyzer_config"`
	Log               *conf.Log                     `toml:"server_log_config"`
	MysqlAccessConfig *conf.RemoteMysqlAccessConfig `toml:"remote_mysql_access"`
	MonitorDBConfig   *conf.MetaDBConfig            `toml:"monitor_db_config"`
	CMDBConfig        *conf.CMDBConfig              `toml:"cmdb_config"`
}

type Basic struct {
	ENV string `toml:"env"`
}

// LoadConfig load config into *Config
func LoadConfig(path string) (config *Config, err error) {
	config = &Config{}
	if err = configor.Load(config, path); err != nil {
		return nil, err
	}

	config.Analyzer.Pattern = slowPattern
	return config, nil
}

type Analyzer struct {
	KeepHint bool `toml:"keep_hint"`
	Pattern  map[string]string
}
