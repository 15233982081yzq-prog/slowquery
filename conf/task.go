package conf

type DailyRankConfig struct {
	Top     int    `toml:"top"`
	OrderBy string `toml:"order_by"`
}
