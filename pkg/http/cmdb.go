package http

import (
	conf "smart-slowquery/conf/platform"
	cmdbUtil "smart-slowquery/internal/util/cmdb"

	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/uic"

	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const RoleAdmin = "admin"
const RoleNormal = "normal"

func (api *Api) GetServiceTree(c *gin.Context) {
	var (
		isAdmin      bool
		err          error
		operator     string
		cmdbServices []string
	)

	operator = api.GetOperatorEmail(c)

	if isAdmin, err = uic.IsDBaasAdmin(c, operator); isAdmin { // DBA Admin
		log.Warningf("http GetServiceTree operator:%s is dbaas admin!", operator)
		response.ToResponse(c, map[string]string{}, fmt.Errorf("dbaas admin unnecessary GetServiceTree"))
		return
	} else if err != nil {
		log.Errorf("http GetServiceTree operator:%s uic.IsDBaasAdmin error:%s", operator, err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	cmdbServices, err = cmdbUtil.GetServiceTree(c, c.Request.Header.Get("Authorization"), "shopee.", conf.GlobalConfig.SpaceConfig.SpaceHost)

	response.ToResponse(c, response.GetServiceTreeResponse{CmdbServices: cmdbServices}, err)
}

func (api *Api) GetUserRole(c *gin.Context) {
	var (
		role    = RoleNormal
		isAdmin bool
		err     error
	)

	operator := api.GetOperatorEmail(c)

	if isAdmin, err = uic.IsDBaasAdmin(c, operator); isAdmin {
		role = RoleAdmin
	}

	response.ToResponse(c, response.GetUserRoleResponse{
		Email: operator,
		Role:  role,
	}, err)
}

func (api *Api) GetLogicDBByService(c *gin.Context) {
	var (
		err         error
		cmdb, dbEnv string
		dbs         []string
	)

	cmdb = c.DefaultQuery("cmdb_service", "")
	dbEnv = c.DefaultQuery("db_env", "")

	if len(dbEnv) == 0 || len(cmdb) == 0 {
		log.Errorf("http GetLogicDBByService param failed,service:%s,dbEnv:%s", cmdb, dbEnv)
		response.ToResponse(c, map[string]string{}, fmt.Errorf("GetLogicDBByService cmdb_service/db_type/db_env is empty"))
		return
	}

	if dbs, err = api.dataBaseSrv.GetDataBases(cmdb, api.GetToken(c), strings.Split(dbEnv, ",")); err != nil {
		log.Errorf("http GetLogicDBByService cmdbUtil.ListLogicalDBService failed,cmdb_service:%s,db_env:%s ,error:%s", cmdb, dbEnv, err.Error())
	}

	response.ToResponse(c, response.GetLogicDBByServiceResponse{LogicDbs: dbs}, err)
}
