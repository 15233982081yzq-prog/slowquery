package http

import (
	"github.com/gin-gonic/gin"

	"smart-slowquery/internal/util/http"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"
	"smart-slowquery/thrid-party/soar-dev/advisor"
)

// 当前这个和Api在同一个http包 所以可以调用他的方法
func (api *Api) Optimize(c *gin.Context) {
	//变量块 多个变量
	var (
		err                                          error
		traceID                                      = c.Value(http.CtxRequestId).(string)
		req                                          = &request.OptimizeRequest{} //req里面通过Bind 把富含SQL信息的c填充到自己的字段
		mysqlSuggest, indexSuggest, heuristicSuggest map[string]advisor.Rule
	)
	//在var req是一个空的结构体 前端传来的json会使用Gin框架 自动绑定到这个结构体 包括sql db等等 如果缺失字段 进入{}
	if err = c.Bind(req); err != nil {
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	// 获取数据库实例信息 调用dbms的接口 获取从库域名信息
	req.Domains, err = api.dbmsSrv.GetSlaveDomain(req.CurrentDB, req.Env, req.ClusterUUID, traceID)
	if err != nil {
		log.Errorf("query detail db:%v, uuid:%v, error:%s, trace_id:%s", req.CurrentDB, req.ClusterUUID, err.Error(), traceID)
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	//数据库访问凭证
	req.User, req.Pass = api.accessSrv.UserName, api.accessSrv.Password
	//获取建议，当前这个和Api在同一个http包 所以可以调用他的方法api的optimizeSrv服务的GetSuggests方法
	mysqlSuggest, indexSuggest, heuristicSuggest, err = api.optimizeSrv.GetSuggests(req, traceID)
	if err != nil {
		log.Errorf("get suggest err:%s, db:%v, uuid:%v, error:%s, trace_id:%s", err, req.CurrentDB, req.ClusterUUID, err.Error(), traceID)
		response.ToResponse(c, map[string]string{}, err)
		return
	}
	//建议包装为响应结构体
	suggestVo := &response.SuggestInfoVo{
		HeuristicSuggest: response.ToRuleVo(heuristicSuggest, response.WithOutCase()),
		MysqlSuggest:     response.ToRuleVo(mysqlSuggest),
		IndexSuggest:     response.ToRuleVo(indexSuggest),
	}

	response.ToResponse(c, suggestVo, nil)
}
