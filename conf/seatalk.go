package conf

type SeatalkRobotConfig struct {
	DbaRobots []string `toml:"dba_robots"`
	DevRobots []string `toml:"dev_robots"`
}
