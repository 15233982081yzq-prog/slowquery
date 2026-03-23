package dbms

import (
	"errors"
	"fmt"
	"time"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	"smart-slowquery/pkg/service/cmdb"
)

var targetsRole = []string{"shadow", "bislave", "backendslave", "master"}

type DBMetaService struct {
	cfg *conf.CMDBConfig
}

func NewDbMetaService(cmdbCfg *conf.CMDBConfig) (*DBMetaService, error) {
	return &DBMetaService{cfg: cmdbCfg}, nil
}

func (meta *DBMetaService) GetOwnCMDB(dbName, dbEnv string) (ownCMDB string, err error) {
	var details *cmdb.GetDatabaseDetailResponse

	log.Infof("GetOwnCMDB spaceHost:%s,db:%s,env:%s", meta.cfg.SpaceHost, dbName, dbEnv)
	if details, err = cmdb.GetDataBaseDetail(dbName, dbEnv, meta.cfg.DBAMetaAuthToken, meta.cfg.SpaceHost); err != nil {
		log.Warningf("cmdb GetOwnCMDB.GetDataBaseDetail dbName:%s ,env:%s ,error:%s", dbName, dbEnv, err.Error())
		return "", err
	}
	if details == nil {
		log.Warningf("cmdb GetOwnCMDB.GetDataBaseDetail dbName:%s ,env:%s ,error:%s", dbName, dbEnv, err.Error())
		return "", fmt.Errorf("not found cmdbservice metadata")
	}

	return details.GetOwnerCmdb(), nil
}

func (meta *DBMetaService) GetL1L2AndTeamAndRoleInfo(dbName, dbEnv string) (l1l2 string, team string, roleMap map[string]string, err error) {
	var details *cmdb.GetDatabaseDetailResponse

	log.Infof("GetOwnCMDB spaceHost:%s,db:%s,env:%s", meta.cfg.SpaceHost, dbName, dbEnv)
	if details, err = cmdb.GetDataBaseDetail(dbName, dbEnv, meta.cfg.DBAMetaAuthToken, meta.cfg.SpaceHost); err != nil {
		log.Warningf("cmdb GetOwnCMDB.GetDataBaseDetail dbName:%s ,env:%s ,error:%s", dbName, dbEnv, err.Error())
		return "", "", nil, err
	}
	if details == nil {
		log.Warningf("cmdb GetOwnCMDB.GetDataBaseDetail dbName:%s ,env:%s ,error:%s", dbName, dbEnv, errors.New("dbms resp is nil"))
		return "", "", nil, fmt.Errorf("not found cmdbservice metadata")
	}

	return details.GetL1L2(), details.GetTeam(), details.GetRoleMap(), nil
}

func (meta *DBMetaService) GetOwnShip(dbName, dbEnv string) (cmdbService, productLine, owner string, leaders []string, err error) {
	var details *cmdb.GetDatabaseDetailResponse

	log.Infof("GetOwnShip spaceHost:%s,db:%s,env:%s", meta.cfg.SpaceHost, dbName, dbEnv)
	if details, err = cmdb.GetDataBaseDetail(dbName, dbEnv, meta.cfg.DBAMetaAuthToken, meta.cfg.SpaceHost); err != nil {
		log.Warningf("cmdb GetOwnShip.GetDataBaseDetail dbName:%s ,env:%s ,error:%s", dbName, dbName, err.Error())
		return "", "", "", nil, err
	}

	cmdbService, productLine, owner, leaders = details.GetL2OwnerShip()
	return
}

func (meta *DBMetaService) GetSlaveDomain(dbName, env, clusterUUID, sourceType string) (domains cmdb.Domains, err error) {
	var details *cmdb.GetDatabaseDetailResponse

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("DBMetaService.GetSlaveDomain", sysMetrics.GetStatus(err), time.Since(start))

	log.Infof("GetSlaveDomain spaceHost:%s,db:%s,env:%s", meta.cfg.SpaceHost, dbName, env)
	if details, err = cmdb.GetDataBaseDetail(dbName, env, meta.cfg.DBAMetaAuthToken, meta.cfg.SpaceHost); err != nil {
		log.Warningf("cmdb GetSlaveDomain.GetDataBaseDetail dbName:%s ,env:%s ,error:%s", dbName, env, err.Error())
		return nil, err
	}

	domainList := details.GetDomainsByClusterUUID(clusterUUID)
	if len(domainList) == 0 {
		return nil, fmt.Errorf("not found shadow/bislave/backendslave domain")
	}

	for _, target := range targetsRole {
		if targetDomain := selectDomain(domainList, target); targetDomain != nil {
			domains = append(domains, targetDomain)
		}
	}

	return domains, nil
}

func selectDomain(domains []*cmdb.Domain, target string) *cmdb.Domain {
	for _, domain := range domains {
		if domain.Role == target {
			return domain

		}
	}
	return nil
}
