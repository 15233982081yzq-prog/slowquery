package filter

import (
	"net/http"
	"strings"

	"git.garena.com/shopee/go-shopeelib/gin/middlewares"
	"git.garena.com/shopee/go-shopeelib/spacelib/models/auth"
	"github.com/gin-gonic/gin"

	conf "smart-slowquery/conf/alert"
	"smart-slowquery/pkg/log"
)

func AlertApiAuthMiddleware(spaceHost string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			email string
			user  *auth.Auth
		)
		spaceToken, err := middlewares.GetSpaceToken(c)
		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				&Response{
					Success: false,
					Result:  nil,
					Error:   err.Error(),
				})
			c.Abort()
			return
		}
		if strings.Contains(c.Request.URL.Path, "monitor/alert_callback") {
			if spaceToken != conf.GlobalConfig.CallBackToken {
				log.Infof("AlertApiAuthMiddleware ,spaceToken:%s ,GlobalConfig.OpenApiToken:%s", spaceToken, conf.GlobalConfig.CallBackToken)
				c.JSON(
					http.StatusUnauthorized,
					&Response{
						Success: false,
						Result:  nil,
						Error:   "slow query alert server auth 401",
					})
				c.Abort()
			} else {
				c.Next()
			}
		} else {
			handler := auth.NewAuthHandler()
			if user, err = handler.LoginByAuthorizationWithURL(spaceHost, spaceToken); err != nil {
				c.JSON(
					http.StatusUnauthorized,
					&Response{
						Success: false,
						Result:  nil,
						Error:   err.Error(),
					})
				c.Abort()
				return
			}
			email = user.User.Email
			c.Set("EMAIL", email)
			c.Next()
		}
	}
}
