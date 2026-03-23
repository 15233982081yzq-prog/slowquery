package cmdb

import (
	conf2 "smart-slowquery/conf"
	conf "smart-slowquery/conf/platform"
	"smart-slowquery/pkg/log"

	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	cmdb = "shopee.engineering_infra.infra_products.db_products.db_dns_qa_test"
)

func TestGetUrl(t *testing.T) {
	srv, err := initTestContext(dir)
	assert.NoError(t, err, "init cmdb_database_service error")

	assert.NotEmpty(t, srv.getUrl(cmdb))
}

func TestName(t *testing.T) {
	srv, _ := initTestContext(dir)
	srv.GetDataBases("", "", nil)
	srv.GetDataBases("tmp", "", nil)
}

func initTestContext(dir string) (service *DataBasesService, err error) {
	var cfg *conf.Config

	if cfg, err = initConfig(dir); err != nil {
		return nil, err
	}
	log.InitLog(cfg.ServerLog)
	NewDataBaseService(nil)
	NewDataBaseService(&conf2.CMDBDataBase{
		Space:  "",
		Path:   "",
		Method: "",
	})
	return NewDataBaseService(cfg.CMDBDataBaseConfig)
}

func initConfig(dir string) (*conf.Config, error) {
	cfg, err := conf.LoadConfig(dir)
	return cfg, err
}
