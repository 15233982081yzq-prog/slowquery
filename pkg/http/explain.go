package http

import (
	"fmt"
	"time"

	"smart-slowquery/internal/util/env"
	"smart-slowquery/internal/util/errors"
	"smart-slowquery/internal/util/hint"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	mysqlParser "smart-slowquery/pkg/parser/mysql"

	storeMysql "smart-slowquery/pkg/store/mysql"
	store "smart-slowquery/pkg/store/response"
	storeResp "smart-slowquery/pkg/store/response"

	"github.com/gin-gonic/gin"
)

func (api *Api) QueryExplain(c *gin.Context) {

	var (
		err            error
		sqlType        string
		explainRequest = &request.ExplainRequest{}
		result         *storeResp.ExplainResult
	)

	if err = BindJsonParam(c, explainRequest); err != nil {
		log.Errorf("http explain param failed, error:%s \n", err.Error())
		response.ToAbortErrorResponse(c, errors.AnnotateParameterErrorf(err, "http request error"))
		return
	}
	if sqlType, err = mysqlParser.ParseSqlStatement(hint.RemoveHint(explainRequest.SQL)); err != nil {
		log.Errorf("get statement type err, sql :%s,error:%s", explainRequest.SQL, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	if sqlType == mysqlParser.InsertStatement {
		log.Errorf("not support insert sql to exec explain, sql:%s,", explainRequest.SQL)
		response.ToResponse(c, map[string]string{}, fmt.Errorf("not support insert sql to exec explain, sql:%s", explainRequest.SQL))
		return
	}

	explainRequest.User, explainRequest.Pass = api.accessSrv.UserName, api.accessSrv.Password

	explainRequest.Domains, err = api.dbmsSrv.GetSlaveDomain(explainRequest.DBName, explainRequest.Env, explainRequest.ClusterUUID, env.DBTypeMySQL)
	if err != nil {
		log.Errorf("query detail domainUrl:%v,error:%s", explainRequest.Domains, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	result, err = fetchExplain(explainRequest)
	if err != nil {
		log.Errorf("exec explain err, request is :%v,error:%s", explainRequest, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	response.ToResponse(c, response.BuildExplainVo(result), err)
}

func fetchExplain(explainRequest *request.ExplainRequest) (*storeResp.ExplainResult, error) {
	var (
		err     error
		s       *storeMysql.Session
		results []*store.ExplainInfo
	)

	start := time.Now()
	defer sysMetrics.CollectServiceMetrics("fetchExplain", sysMetrics.GetStatus(err), time.Since(start))

	for _, slaveDomain := range explainRequest.Domains {
		if s, err = storeMysql.NewSession(explainRequest.User, explainRequest.Pass, explainRequest.DBName, slaveDomain.Domain, int32(slaveDomain.Port)); err != nil {
			log.Errorf("fetchExplain NewSession error:%s", err.Error())
			_ = s.Close()
			continue
		}
		if results, err = s.ExplainWithTimeout(explainRequest.SQL, 10*time.Second); err != nil {
			log.Errorf("fetchExplain ExplainWithTimeout statement:%s ,session:%s,error:%s", explainRequest.SQL, s.String(), err.Error())
		}
		_ = s.Close()
		if len(results) > 0 {
			break
		}
	}
	return store.NewExplainResult(results, err), err
}
