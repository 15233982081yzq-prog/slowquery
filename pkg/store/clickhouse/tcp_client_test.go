package clickhouse

func initDatabaseReadWrite(dbSetting *TestDBSetting) (*Client, error) {
	// because write use TCP protocol
	dbSetting.Config.Port = clickhouseTCPPort
	defer func() {
		dbSetting.Config.Port = clickhouseHTTPPort
	}()
	return NewTcpClient(dbSetting.Config)
}
