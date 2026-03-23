package conf

import (
	"fmt"
	"time"
)

type CMDBDataBase struct {
	Space   string   `toml:"space"`
	Path    string   `toml:"path"`
	Method  string   `toml:"http_method"`
	Timeout duration `toml:"client_timeout"`
}

func (cmdb *CMDBDataBase) Validate() error {
	if len(cmdb.Space) == 0 || len(cmdb.Path) == 0 || len(cmdb.Method) == 0 {
		return fmt.Errorf("Space config failed, Space:%s,Path:%s ,Method:%s \n", cmdb.Space, cmdb.Path, cmdb.Method)
	}
	return nil
}

type CMDBConfig struct {
	SpaceHost        string `toml:"space_host"`
	DBAMetaAuthToken string `toml:"dbameta_auth_token"`
}

// support toml set time.duration
type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
