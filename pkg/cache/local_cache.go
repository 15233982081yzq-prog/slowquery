package cache

import (
	"github.com/allegro/bigcache"

	"smart-slowquery/conf"
)

type LocalCache struct {
	cache *bigcache.BigCache
	cfg   *conf.LocalCache
}

func NewLocalCache(cfg *conf.LocalCache) (*LocalCache, error) {
	config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: cfg.Shards,

		// time after which entry can be evicted
		LifeWindow: cfg.ExpireTime.Duration,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: cfg.CleanTime.Duration,

		// prints information about additional memory allocation
		Verbose: true,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: cfg.MaxCacheSize,
	}

	var (
		cache *bigcache.BigCache
		err   error
	)

	if cache, err = bigcache.NewBigCache(config); err != nil {
		return nil, err
	}

	return &LocalCache{
		cache: cache,
		cfg:   cfg,
	}, nil
}

func (lc *LocalCache) Get(key string) ([]byte, error) {
	return lc.cache.Get(key)
}

func (lc *LocalCache) Set(key string, entry []byte) error {
	return lc.cache.Set(key, entry)
}

func (lc *LocalCache) Stats() bigcache.Stats {
	return lc.cache.Stats()
}
