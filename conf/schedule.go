package conf

import "time"

type SchedulerConfig struct {
	Cycle    time.Duration `toml:"cycle"`
	Parallel int           `toml:"parallel"`
}
