package conf

import "time"

type DodConfig struct {
	DODBaseURL string        `toml:"dod_base_url"`
	Timeout    time.Duration `toml:"timeout"`
}
