package filter

import (
	httpUtil "smart-slowquery/internal/util/http"

	"github.com/gin-gonic/gin"

	uuid "github.com/satori/go.uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start := time.Now()
		requestID := c.Request.Header.Get("X-Request-Id")
		if requestID == "" {
			newUUID := uuid.NewV4()
			requestID = newUUID.String()
		}
		c.Set(httpUtil.CtxRequestId, requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}
