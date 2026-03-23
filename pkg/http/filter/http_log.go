package filter

import (
	httpUtil "smart-slowquery/internal/util/http"

	"smart-slowquery/pkg/log"

	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

var running = atomic.NewInt64(0)

// for log gin response
// https://stackoverflow.com/questions/38501325/how-to-log-response-body-in-gin
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		running.Add(1)
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 请求包写回
		reqBody, _ := c.GetRawData()
		if len(reqBody) > 0 {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 响应包写回
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()

		allCost := time.Since(start).Milliseconds() // 本次请求的总共消耗时间
		externalCost := c.GetInt64(httpUtil.CtxExternalCost)
		running.Add(-1)
		// 写入日志
		msg := "Http Request Info:"
		msg += " path=" + path
		msg += " request-id=" + c.GetString(httpUtil.CtxRequestId)
		msg += " method=" + c.Request.Method
		msg += " ip=" + c.ClientIP()
		msg += " user-agent=" + c.Request.UserAgent()
		msg += " errors=" + c.Errors.ByType(gin.ErrorTypePrivate).String()
		msg += fmt.Sprintf(" all-cost=%dms", allCost)
		msg += fmt.Sprintf(" external-cost=%dms", externalCost)
		msg += fmt.Sprintf(" self-cost=%dms", allCost-externalCost)
		msg += " query=" + query
		msg += " request-body=" + string(reqBody)
		msg += fmt.Sprintf(" status=%d", c.Writer.Status())
		msg += " response-body=" + blw.body.String()

		log.Info(msg)
	}
}

func HttpRunning() int64 {
	return running.Load()
}
