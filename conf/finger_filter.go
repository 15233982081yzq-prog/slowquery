package conf

type FingerFilter struct {
	Satisfaction float64 `toml:"satisfaction"`
	DefaultCap   int     `toml:"default_cap"`
}

type FreshFingerFilter struct {
	CountBySingleGoRoutine int      `toml:"count_by_single"`
	MaxConcurrent          int      `toml:"max_concurrent"`
	BatchSize              int      `toml:"batch_size"`
	CronSpec               string   `toml:"cron_spec"`
	ValidDays              int      `toml:"valid_days"`
	QueryLeadTime          duration `toml:"query_lead"`
}
