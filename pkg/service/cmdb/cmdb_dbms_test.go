package cmdb

import (
	platformConf "smart-slowquery/conf/platform"
	"smart-slowquery/pkg/log"

	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dir       = "./test/cmdb-database-config.test.toml"
	spaceHost string
	token     string

	dbName = "shopee_credit_broker_uat_cn_db"
	env    = "live"
)

func TestGetDataBaseDetail(t *testing.T) {
	databaseDetail, err := GetDataBaseDetail(dbName, env, token, spaceHost)
	assert.NoError(t, err)
	assert.NotNil(t, databaseDetail)
}

func init() {
	config, _ := platformConf.LoadConfig(dir)
	token = config.CMDBConfig.DBAMetaAuthToken
	spaceHost = config.CMDBConfig.SpaceHost
	log.InitLog(config.ServerLog)
}

func TestGetServiceTree(t *testing.T) {
	GetServiceTree(token, spaceHost, "")
}
