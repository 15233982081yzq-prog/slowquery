package alert

import (
	"time"

	storeMsql "smart-slowquery/pkg/store/mysql"
)

type ChannelTab struct {
	ID          int       `gorm:"column:id"`
	ChannelUUID string    `gorm:"column:channel_uuid;uniqueIndex"`
	ChannelName string    `gorm:"column:channel_name"`
	DodID       int       `gorm:"column:dod_id"`
	UsersJson   string    `gorm:"column:users_json"`
	Interval    string    `gorm:"column:channel_interval"`
	MetaJson    string    `gorm:"column:meta_json"`
	Status      string    `gorm:"column:channel_status"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

func (ct *ChannelTab) TableName() string {
	return "alert_channel_tab"
}

func (ct *ChannelTab) Save(db storeMsql.DB) error {
	return db.Save(ct)
}

func (ct *ChannelTab) UpdateByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&ChannelTab{}).Where(ChannelTab{ChannelUUID: ct.ChannelUUID}).Updates(ct).Error
}

func (ct *ChannelTab) UpdateStatusByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&ChannelTab{}).Where(ChannelTab{ChannelUUID: ct.ChannelUUID}).Updates(&ChannelTab{Status: ct.Status}).Error
}

func (ct *ChannelTab) DeleteByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&ChannelTab{}).Where(ChannelTab{ChannelUUID: ct.ChannelUUID}).Delete(&ChannelTab{}).Error
}

func FindAlertChannelByUUID(db storeMsql.DB, uuid string) (ar *ChannelTab, err error) {
	res := &ChannelTab{}
	if err = db.GetDBConn().Model(&ChannelTab{}).Where(ChannelTab{ChannelUUID: uuid}).First(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
