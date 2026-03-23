package conf

type CKCli struct {
	Driver      string   `toml:"driver"`
	Host        string   `toml:"host"`
	Port        int32    `toml:"port"`
	User        string   `toml:"user"`
	Password    string   `toml:"password"`
	DbName      string   `toml:"db_name"`
	Compression bool     `toml:"compression"`
	MaxOpenConn int      `toml:"max_open_conn"`
	MaxIdleConn int      `toml:"max_idle_conn"`
	MaxLifeTime duration `toml:"max_life_time"`
}

type CKFlusher struct {
	Cycle duration `toml:"cycle"`
	Batch int      `toml:"batch"`
}
