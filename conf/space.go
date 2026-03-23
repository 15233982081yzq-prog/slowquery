package conf

import (
	"fmt"
)

type Space struct {
	SpaceHost string `toml:"space_host"`
	SpaceEnv  string `toml:"space_env"`
	User      string `toml:"user"`
	Pass      string `toml:"pass"`
}

func (sp *Space) Validate() error {
	if len(sp.SpaceHost) == 0 || len(sp.User) == 0 || len(sp.Pass) == 0 {
		return fmt.Errorf("Space config failed, SpaceHost:%s,User:%s ,Pass:%s \n", sp.SpaceHost, sp.User, sp.Pass)
	}
	return nil
}

type UserGroupConfig struct {
	DBassTeamId uint64 `toml:"dbass_team_id"`
	DBADodId    uint64 `toml:"dba_dod_id"`
	TestMode    bool   `toml:"test_mode"`
}
