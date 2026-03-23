package uic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/cmdb"

	bashConf "smart-slowquery/conf"
	httpUtil "smart-slowquery/internal/util/http"
	stringUtil "smart-slowquery/internal/util/string"
	sysMetrics "smart-slowquery/pkg/metrics/platform"

	"git.garena.com/shopee/platform/space-sdk/uic"
	"github.com/gin-gonic/gin"
)

var (
	uicClient     uic.ClientInterface
	userGroupInfo *bashConf.UserGroupConfig
)

func InitUicClient(spaceConf *bashConf.Space, userGroupConf *bashConf.UserGroupConfig) (err error) {
	if spaceConf == nil || userGroupConf == nil {
		return errors.New("conf is nil")
	}
	uicClient, err = cmdb.NewUicClient(spaceConf)
	if err != nil {
		return err
	}
	userGroupInfo = userGroupConf
	return nil
}

func getMemberList(c *gin.Context, teamId uint64) ([]string, error) {
	if uicClient == nil {
		return nil, errors.New("uicClient is nil, please first use InitUicClient to init client")
	}
	start := time.Now()
	defer func() {
		cost := time.Since(start).Milliseconds()
		r := c.GetInt64(httpUtil.CtxExternalCost)
		c.Set(httpUtil.CtxExternalCost, cost+r)
		log.Infof("cmdb cost.getMemberList=%v, request_id=%v", cost, c.GetString(httpUtil.CtxRequestId))
	}()

	res, err := uicClient.GetGroupMembers(context.Background(), uic.GetGroupMemberReq{
		GroupIDs: []uint64{teamId},
	})
	if err != nil {
		return nil, err
	}

	if !res.Success {
		return nil, fmt.Errorf("get UIC list error, teamid = %v", teamId)
	}

	var ul []string
	for _, group := range res.GroupMembers {
		for _, user := range group.Members {
			ul = append(ul, user.Email)
		}
	}
	return ul, nil
}

func IsDBaasAdmin(c *gin.Context, user string) (bool, error) {
	var (
		err     error
		members []string
		start   = time.Now()
	)
	if uicClient == nil {
		return false, errors.New("uicClient is nil, please first use InitUicClient to init client")
	}

	sysMetrics.CollectServiceMetrics("uic", sysMetrics.GetStatus(err), time.Since(start))

	if userGroupInfo.TestMode {
		return true, nil
	}

	if members, err = getMemberList(c, userGroupInfo.DBassTeamId); err != nil {
		log.Errorf("user group get failed, please contact admin , error:%s", err.Error())
		return false, err
	}

	return stringUtil.ContainInSlice(members, user), nil
}
