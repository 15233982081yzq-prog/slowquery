package dbms

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"smart-slowquery/conf"
	platformConf "smart-slowquery/conf/platform"

	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/cmdb"
)

var (
	dir            = "./test/dbms-config.test.toml"
	logicDBsDir    = "./test/logic_dbs.txt"
	spaceHost      string
	token          string
	logicDBs       []string
	env            = "live"
	clusterUUIDMap = make(map[string][]string, 0)
	config         *platformConf.Config
	dbMetaService  *DBMetaService
)

func TestGetDataBaseDetail(t *testing.T) {
	for _, dbName := range logicDBs {
		//fmt.Printf("dbname:%s,env:%s \n", dbName, env)
		databaseDetail, err := cmdb.GetDataBaseDetail(dbName, env, token, spaceHost)
		if err != nil {
			fmt.Printf("%s ,resource_not_exists\n", dbName)
			continue
		}
		//assert.NoError(t, err)
		//assert.NotNil(t, databaseDetail)
		uuids := databaseDetail.GetClusterUUIDs()
		for _, uuid := range uuids {
			dbs := clusterUUIDMap[uuid]
			dbs = append(dbs, dbName)
			clusterUUIDMap[uuid] = dbs
		}
	}
	for k, v := range clusterUUIDMap {
		fmt.Printf("%s ,%v \n", k, v)
	}
}

func init() {
	var err error
	config, _ = platformConf.LoadConfig(dir)
	token = config.CMDBConfig.DBAMetaAuthToken
	spaceHost = config.CMDBConfig.SpaceHost
	log.InitLog(config.ServerLog)
	if logicDBs, err = readLogicDbsFile(logicDBsDir); err != nil {
		fmt.Printf("dbmeta_test init error:%s", err.Error())
	}
	// Init log, tem config
	_ = log.InitLog(&conf.Log{
		Name:        "test",
		Level:       "info",
		Path:        "./tmp/log",
		MaxFileSize: 500,
		MaxBackups:  10,
		MaxAge:      10,
		Compress:    true,
	})
	dbMetaService, err = NewDbMetaService(config.CMDBConfig)
	if err != nil {
		return
	}
}

func TestDBMetaService_GetOwnCMDB(t *testing.T) {
	dbMetaService.GetOwnCMDB("shopee_bianjian_slow_pic5_db", "live")
	dbMetaService.GetOwnCMDB("shopee_bianjian_slow_pic5_db_none", "live")
	dbMetaService.GetOwnShip("shopee_bianjian_slow_pic5_db", "live")
	dbMetaService.GetOwnShip("shopee_bianjian_slow_pic5_db_none", "live")
}

func readLogicDbsFile(path string) (dbs []string, err error) {
	var f *os.File
	// open file
	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file word by word using scanner
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		// do something with a word
		dbs = append(dbs, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return
}

func TestDBMetaService_GetSlaveInstance(t *testing.T) {
	meta := &DBMetaService{
		cfg: &conf.CMDBConfig{
			SpaceHost:        spaceHost,
			DBAMetaAuthToken: token,
		},
	}

	meta.GetSlaveDomain("shopee_bianjian_slow_pic5_db", env, "8ce3c5b8355dd223", "test_trace_id")
	meta.GetSlaveDomain("shopee_bianjian_slow_pic5_db", env, "100000209641", "test_trace_id")

}

func TestSelectDomain(t *testing.T) {
	cmdbs := []*cmdb.Domain{
		{
			Domain:     "123",
			Port:       0,
			DomainType: "123",
			Role:       "master",
		},
	}
	assert.NotEqual(t, nil, selectDomain(cmdbs, "master"))
}
