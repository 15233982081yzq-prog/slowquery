package conf

type LocalCache struct {
	Shards       int      `toml:"shards"`
	ExpireTime   duration `toml:"expireTime"`
	CleanTime    duration `toml:"clean_time"`
	MaxCacheSize int      `toml:"max_cache_size"`
}
