package http

import (
	"smart-slowquery/internal/util/errors"
	"smart-slowquery/internal/util/sys"
	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"

	"github.com/gin-gonic/gin"
)

func (api *Api) SetRemoteDBPasswd(c *gin.Context) {
	var err error

	req := &request.SetRemoteDBPasswdRequest{}
	if err = BindJsonParam(c, req); err != nil {
		log.Errorf("http SetRemoteDBPasswd param failed, error:%s \n", err.Error())
		response.ToAbortErrorResponse(c, errors.AnnotateParameterErrorf(err, "http request error"))
		return
	}

	log.Infof("user %v call SetRemoteDBPasswd", api.GetOperatorEmail(c))

	if err = api.accessSrv.CreateOrUpdateMySQLAccess(req.UserName, req.Password); err != nil {
		log.Errorf("http SetRemoteDBPasswd accessSrv.CreateOrUpdateMySQLAccess failed, error:%s \n", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	response.ToResponse(c, req, nil)
}

func (api *Api) GetRemoteDBPasswd(c *gin.Context) {
	var err error

	log.Infof("user %v call GetRemoteDBPasswd", api.GetOperatorEmail(c))

	if err = api.accessSrv.GetPasswdHash(); err != nil {
		log.Errorf("http GetRemoteDBPasswd accessSrv.UpdatePasswdHash() failed, error:%s \n", err.Error())
		response.ToResponse(c, map[string]string{}, err)
		return
	}

	p, k, m := api.accessSrv.GetAccessMeta()

	response.ToResponse(c, response.SetRemoteDBPasswdResponse{
		UserName:   m.UserName,
		Password:   p,
		Key:        k,
		UpdateTime: m.UpdateTime.Format(sys.TimeNormalFormat),
	}, nil)
}
