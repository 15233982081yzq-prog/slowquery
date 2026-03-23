package mysql

import "time"

type AccessMeta struct {
	Id           int64     `gorm:"id"`
	UserName     string    `gorm:"user_name"`
	Env          string    `gorm:"env"`
	PasswordHash string    `gorm:"password_hash"`
	UpdateTime   time.Time `gorm:"update_time;autoCreateTime;autoUpdateTime"`
	//CreateTime   time.Time `gorm:"create_time;autoCreateTime"`
}

func (meta *AccessMeta) TableName() string {
	return "internal_user_tab"
}
