package cmdb

import (
	stringUtil "smart-slowquery/internal/util/string"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"

	"context"
	"fmt"
	"strings"

	"git.garena.com/shopee/platform/space-sdk/core/space"
)

const (
	dbmetaNoPermission = "no permission to view dbameta databases details by service"
)

type Request struct {
	CmdbService string `json:"service_name"`
}

type DBMeta struct {
	Uuid         string `json:"uuid"`
	DatabaseName string `json:"database_name"`
	DatabaseType string `json:"database_type"`
	Environment  string `json:"environment"`
	IsRetired    bool   `json:"is_retired"`
	Instances    []struct {
		Role    string   `json:"role"`
		Domains []string `json:"domains"`
	} `json:"instances"`
	DrConfig         string `json:"dr_config"`
	DbIdentification string `json:"db_identification"`
	DbMarket         string `json:"db_market"`
}

type GetResponse struct {
	Databases    []*DBMeta `json:"databases"`
	BusinessCode int       `json:"business_code"`
	Success      bool      `json:"success"`
}

type DataBasesService struct {
	getDataBaseConf *conf.CMDBDataBase
}

func NewDataBaseService(config *conf.CMDBDataBase) (*DataBasesService, error) {
	if config == nil {
		return nil, fmt.Errorf("init cmdb_database_service failed config is empty")
	}

	if config.Validate() != nil {
		return nil, fmt.Errorf("init cmdb_database_service failed config invalid,error:%v", config.Validate())
	}

	return &DataBasesService{
		getDataBaseConf: config,
	}, nil
}

func (srv *DataBasesService) GetDataBases(cmdb, token string, envs []string) (logicDBs []string, err error) {
	log.Infof("cmdb dataBasesService GetDataBases cmdb:%s", cmdb)

	if len(cmdb) == 0 {
		return nil, fmt.Errorf("databaseService getDataBase param failed! cmdb:%s", cmdb)
	}

	var getResp GetResponse
	if err = space.NewHTTPClient().Get(context.TODO(), srv.getUrl(cmdb), nil, &getResp, space.WithBearerAuthorization(token)); err != nil {
		if strings.Contains(err.Error(), dbmetaNoPermission) {
			err = fmt.Errorf(dbmetaNoPermission)
		}
		return nil, err
	}

	for _, database := range getResp.Databases {
		if database.IsRetired {
			continue
		}

		if database.DatabaseType != "MySQL" {
			continue
		}

		if !stringUtil.ContainInSlice(envs, database.Environment) {
			continue
		}

		if database.DatabaseName == "" {
			continue
		}

		logicDBs = append(logicDBs, database.DatabaseName)
	}

	return logicDBs, nil
}

func (srv *DataBasesService) getUrl(cmdb string) string {
	return fmt.Sprintf("%s%s?service_name=%s", srv.getDataBaseConf.Space, srv.getDataBaseConf.Path, cmdb)
}
