package service

import (
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"

	"fmt"
	"time"
)

type QueryService struct {
	store store.CKStore
}

func NewQueryService(ckStore store.CKStore) (*QueryService, error) {
	if ckStore == nil {
		return nil, fmt.Errorf("ckStore is nil")
	}

	return &QueryService{
		store: ckStore,
	}, nil
}

func (srv *QueryService) GetInstanceHosts(dbName, dbEnv string, startTime, endTime int64) ([]string, error) {
	return srv.store.GetInstanceHosts(request.BuildSlowQueryInstanceHosts(dbName, dbEnv, startTime, endTime))
}

func (srv *QueryService) GetClientTraceability(dbName, dbEnv, fingerID string, instances []string, startTime, endTime int64) (*[]response.ClientHostStats, error) {
	return srv.store.GetClientHostsStats(request.BuildSlowQueryClientHostsStats(dbName, dbEnv, fingerID, instances, startTime, endTime))
}

func (srv *QueryService) GetQueryStatements(dbName, dbEnv, fingerID, orderBy string, instances []string, startTime, endTime int64, offset, limit int) (stmts *[]response.QueryStatement, count int, err error) {
	if count, err = srv.store.GetQueryStatementsCount(request.BuildSlowQueryStatementWithOrderBy(dbName, dbEnv, fingerID, orderBy, instances, startTime, endTime, offset, limit)); err != nil {
		log.Errorf("service GetQueryStatements GetQueryStatementsCount error:%s", err.Error())
		return nil, count, err
	}

	if stmts, err = srv.store.GetQueryStatements(request.BuildSlowQueryStatementWithOrderBy(dbName, dbEnv, fingerID, orderBy, instances, startTime, endTime, offset, limit)); err != nil {
		log.Errorf("service GetQueryStatements GetQueryStatements error:%s", err.Error())
		return stmts, count, err
	}

	return stmts, count, err
}

func (srv *QueryService) GetSlowQueryList(dbEnv, orderBy string, clusterUUIDs, dbNames, instances []string, limit, offset int, startTime, endTime int64, appearType *request.AppearType) (list *[]response.SlowQuery, count int, err error) {

	start := time.Now()
	if count, err = srv.store.GetSlowQueryCount(request.BuildSlowQueryCount(dbEnv, clusterUUIDs, dbNames, instances, startTime, endTime, appearType)); err != nil {
		log.Errorf("service GetSlowQueryList GetSlowQueryCount error:%s", err.Error())
		return nil, -1, err
	}
	log.Infof("QueryService getSlowQueryList db.GetSlowQueryCount cost:%v", time.Since(start))

	if count == 0 {
		log.Warningf("QueryService GetSlowQueryList not found slowQuery message! param database_name=%v ,env=:%s ,instance:%v ,startTime=%d ,endTime=:%d", dbNames, dbEnv, instances, startTime, endTime)
		return &[]response.SlowQuery{}, 0, err
	}

	start = time.Now()
	if list, err = srv.store.GetSlowQueryList(request.BuildSlowQueryList(dbEnv, orderBy, clusterUUIDs, dbNames, instances, limit, offset, startTime, endTime, appearType)); err != nil {
		log.Errorf("service GetSlowQueryList GetSlowQueryList error:%s", err.Error())
		return nil, -1, err
	}
	log.Infof("QueryService getSlowQueryList db.GetSlowQueryList cost:%v", time.Since(start))

	return
}

func (srv *QueryService) FindLast7Days(idList []string) (m map[string]bool, err error) {
	m = make(map[string]bool)
	for _, id := range idList {
		m[id] = false
	}
	start := time.Now()
	list, err := srv.store.GetLast7Days(idList)
	if err != nil {
		log.Errorf("service FindLast7Days FindLast7Days error:%s", err.Error())
		return nil, err
	}
	log.Infof("QueryService FindLast7Days db.FindLast7Days cost:%v, list:%v", time.Since(start), list)
	for _, v := range list {
		m[v] = true
	}
	return
}

func (srv *QueryService) GetQueryDetail(dbName, dbEnv, fingerID string, instances []string, startTime, endTime int64, appearType *request.AppearType) (stmt *response.QueryStatement, users []string, err error) {

	if stmt, err = srv.store.GetQueryStatementOne(request.BuildSlowQueryStatement(dbName, dbEnv, fingerID, instances, startTime, endTime, appearType)); err != nil {
		log.Errorf("service GetQueryDetail GetQueryStatementOne error:%s", err.Error())
		return nil, nil, fmt.Errorf("GetQueryDetail error:%s", err.Error())
	}

	if users, err = srv.store.GetClientUsers(request.BuildSlowQueryClientUsers(dbName, dbEnv, fingerID, instances, startTime, endTime)); err != nil {
		log.Errorf("service GetQueryDetail GetClientUsers error:%s", err.Error())
		return nil, nil, fmt.Errorf("GetQueryDetail error:%s", err.Error())
	}

	return stmt, users, err
}

func (srv *QueryService) GetSlowQueryStatistics(dbEnv string, clusterUUids, dbNames []string, startTime, endTime int64) (list *[]response.SlowQueryDBStatistic, err error) {

	if list, err = srv.store.GetSlowQueryDBStatistics(request.BuildSlowQueryDBStatistic(dbEnv, clusterUUids, dbNames, startTime, endTime)); err != nil {
		log.Errorf("service GetSlowQueryStatistics GetSlowQueryDBStatistics error:%s", err.Error())
		return nil, err
	}

	return
}
