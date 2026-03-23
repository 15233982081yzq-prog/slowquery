package conf

type AlertTemplate struct {
	UUID           string   `toml:"uuid" json:"uuid"`
	Name           string   `toml:"name" json:"name"`
	Expression     []string `toml:"expression" json:"expression"`
	PromQLTemplate string   `toml:"promQL_template" json:"promQL_template"`
	Type           string   `toml:"type" json:"type"`
}
