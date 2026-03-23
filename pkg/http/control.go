package http

import (
	"smart-slowquery/pkg/http/filter"
	"smart-slowquery/pkg/http/response"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/switcher"

	"github.com/gin-gonic/gin"
)

func (api *Api) Server503(c *gin.Context) {
	log.Infof("user:%s call server503", api.GetOperatorEmail(c))

	switcher.CloseServer()
	response.ToResponse(c, map[string]string{"message": "Server503 success"}, nil)
}

func (api *Api) Server200(c *gin.Context) {
	log.Infof("user:%s call server200", api.GetOperatorEmail(c))

	switcher.OpenServer()
	response.ToResponse(c, map[string]string{"message": "Server200 success"}, nil)
}

func (api *Api) ServerRunning(c *gin.Context) {
	log.Infof("user:%s call Running", api.GetOperatorEmail(c))

	body := make(map[string]int64)
	body["http"] = filter.HttpRunning()

	response.ToResponse(c, body, nil)
}

func (api *Api) Switcher(c *gin.Context) {
	log.Infof("user:%s call Platform Switcher", api.GetOperatorEmail(c))

	body := make(map[string]bool)
	body["http"] = switcher.IsServerOpen()

	response.ToResponse(c, body, nil)
}
