package filter

import (
	conf "smart-slowquery/conf/openapi"

	"smart-slowquery/pkg/log"

	"net/http"

	"git.garena.com/shopee/go-shopeelib/gin/middlewares"
	"github.com/gin-gonic/gin"
)

func OpenApiAuthMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if spaceToken, _ := middlewares.GetSpaceToken(c); spaceToken != conf.GlobalConfig.OpenApiToken {
			log.Infof("OpenApiAuthMiddleware version:%s ,spaceToken:%s ,GlobalConfig.OpenApiToken:%s", version, spaceToken, conf.GlobalConfig.OpenApiToken)
			c.JSON(
				http.StatusUnauthorized,
				&Response{
					Version: version,
					Success: false,
					Result:  nil,
					Error:   "slow query server auth 401",
				})
			c.Abort()
		}
		c.Next()
	}
}
