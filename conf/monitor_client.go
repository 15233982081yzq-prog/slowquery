package conf

import "time"

type MonitorClientConfig struct {
	BaseURL          string        `toml:"base_url"`
	Timeout          time.Duration `toml:"timeout"`
	ClientID         string        `toml:"client_id"`
	ClientSecret     string        `toml:"client_secret"`
	MetricStoreNames string        `toml:"metric_store_names"`
	SeatalkMsgID     string        `toml:"seatalk_msg_id"`
	EmailMsgID       string        `toml:"email_msg_id"`
}
