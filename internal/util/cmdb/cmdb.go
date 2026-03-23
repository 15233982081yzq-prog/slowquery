package cmdb

import (
	httpUtil "smart-slowquery/internal/util/http"
	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/cmdb"

	"time"

	"github.com/gin-gonic/gin"
)

func GetServiceTree(c *gin.Context, auth string, name, spaceHost string) ([]string, error) {
	defer deferCostLog(c, "GetServiceTree")()

	start := time.Now()
	res, err := cmdb.GetServiceTree(auth, name, spaceHost)
	if err != nil {
		return []string{}, err
	}
	var li []string
	for _, v := range res.Services {
		if v.ServiceName == "" {
			continue
		}
		li = append(li, v.ServiceName)
	}
	sysMetrics.CollectServiceMetrics("GetServiceTree", sysMetrics.GetStatus(err), time.Since(start))

	return li, nil
}

func deferCostLog(c *gin.Context, funcName string) func() {
	start := time.Now()
	return func() {
		cost := time.Since(start).Milliseconds()
		r := c.GetInt64(httpUtil.CtxExternalCost)
		c.Set(httpUtil.CtxExternalCost, cost+r)
		log.Infof("cmdb cost.%s=%d, request_id=%v", funcName, cost, c.GetString(httpUtil.CtxRequestId))
	}
}
