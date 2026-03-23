package request

import "smart-slowquery/pkg/service/cmdb"

type ExplainRequest struct {
	User        string         `json:"-"`
	Pass        string         `json:"-"`
	Domains     []*cmdb.Domain `json:"-"`
	SQL         string         `json:"sql" binding:"required"`
	DBName      string         `json:"db_name" binding:"required"`
	Env         string         `json:"env" binding:"required"`
	ClusterUUID string         `json:"cluster_uuid" binding:"required"`
}
