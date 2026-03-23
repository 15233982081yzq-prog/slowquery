package request

import (
	storeReq "smart-slowquery/pkg/store/request"

	"smart-slowquery/pkg/service/cmdb"
)

var (
	AppearedMapping = map[string]*storeReq.AppearType{
		storeReq.AppearTypeAll:         storeReq.BuildAppearType(storeReq.AppearTypeAll, -1),
		storeReq.AppearTypeNewAppeared: storeReq.BuildAppearType(storeReq.AppearTypeNewAppeared, 1),
		storeReq.AppearTypeOriginal:    storeReq.BuildAppearType(storeReq.AppearTypeOriginal, 0),
	}
)

type OptimizeRequest struct {
	SQL         string         `json:"sql" binding:"required"`          //需要优化的SQL
	CurrentDB   string         `json:"current_db" binding:"required"`   //数据库名字
	Env         string         `json:"env" binding:"required"`          //Live/NonLive
	ClusterUUID string         `json:"cluster_uuid" binding:"required"` //集群信息
	User        string         `json:"-"`                               //数据库用户名（不暴露前端）
	Pass        string         `json:"-"`                               //数据库密码（不暴露前端）
	Domains     []*cmdb.Domain `json:"-"`                               //数据库实例信息 内部使用
}
