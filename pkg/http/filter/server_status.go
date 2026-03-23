package filter

import (
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/switcher"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HttpStatus(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.FullPath(), "/control") && !switcher.IsServerOpen() {
			log.Infof("HttpStatus switcher.IsServerOpen():%t", switcher.IsServerOpen())
			c.JSON(
				http.StatusServiceUnavailable,
				&Response{
					Version: version,
					Success: false,
					Result:  nil,
					Error:   "slow query server unavailable",
				})
			c.Abort()
		}
		c.Next()
	}
}

type Response struct {
	Version string      `json:"version"`
	Success bool        `json:"success"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}
