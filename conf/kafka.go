package conf

type Kafka struct {
	Topics  []string `toml:"topics"`
	Brokers []string `toml:"brokers"`
	GroupID string   `toml:"groupID"`
}
