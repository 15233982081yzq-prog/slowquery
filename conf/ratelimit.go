package conf

type RateLimitConfig struct {
	Rate     float64 `toml:"rate"`
	Capacity int64   `toml:"capacity"`
}
