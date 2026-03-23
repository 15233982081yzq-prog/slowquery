package conf

type Log struct {
	Name        string `toml:"log_name"`
	Level       string `toml:"log_level"`
	Path        string `toml:"log_path"`
	MaxFileSize int    `toml:"log_max_file_size"` //单位:MB
	MaxBackups  int    `toml:"log_max_backups"`   //文件保存数量
	MaxAge      int    `toml:"log_max_age"`       //文件保存天数
	Compress    bool   `toml:"log_compress"`      //文件压缩
}
