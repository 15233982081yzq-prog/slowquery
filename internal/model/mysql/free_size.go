package mysql

import "time"

type DBFreeSizeTab struct {
	ID          int64     `gorm:"column:id" json:"id"`
	ClusterName string    `gorm:"column:cluster_name" json:"cluster_name"`
	ClusterUUID string    `gorm:"column:cluster_uuid" json:"cluster_uuid"`
	Service     string    `gorm:"column:service" json:"service"`
	DBName      string    `gorm:"column:db_name" json:"db_name"`
	InstRole    string    `gorm:"column:inst_role" json:"inst_role"`
	Domain      string    `gorm:"column:domain" json:"domain"`
	IP          string    `gorm:"column:ip" json:"ip"`
	InstUUID    string    `gorm:"column:inst_uuid" json:"inst_uuid"`
	TotalSize   int64     `gorm:"column:total_size" json:"total_size"`
	FreeSize    int64     `gorm:"column:free_size" json:"free_size"`
	TS          int64     `gorm:"column:ts" json:"ts"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`
	BottomStart string    `gorm:"column:bottom_start" json:"bottom_start"`
	BottomEnd   string    `gorm:"column:bottom_end" json:"bottom_end"`
}

func (d *DBFreeSizeTab) TableName() string {
	return "db_free_size_tab"
}
