package alert

import (
	"time"

	storeMsql "smart-slowquery/pkg/store/mysql"
)

type StrategyTab struct {
	ID             int       `gorm:"column:id"`
	StrategyName   string    `gorm:"column:strategy_name"`
	StrategyID     string    `gorm:"column:strategy_id"`
	MetaJson       string    `gorm:"column:meta_json"`
	StrategyStatus string    `gorm:"column:strategy_status"`
	CreateTime     time.Time `gorm:"column:create_time"`
	UpdateTime     time.Time `gorm:"column:update_time"`
}

func (st *StrategyTab) Save(db storeMsql.DB) error {
	return db.Save(st)
}

func (st *StrategyTab) TableName() string {
	return "alert_strategy_tab"
}

func (st *StrategyTab) UpdateByStrategyID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&StrategyTab{}).Where(StrategyTab{StrategyID: st.StrategyID}).Updates(st).Error
}

func (st *StrategyTab) UpdateByStrategyName(db storeMsql.DB) error {
	return db.GetDBConn().Model(&StrategyTab{}).Where(StrategyTab{StrategyName: st.StrategyName}).Updates(st).Error
}

func (st *StrategyTab) DeleteById(db storeMsql.DB) error {
	return db.GetDBConn().Model(&StrategyTab{}).Where(StrategyTab{StrategyID: st.StrategyID}).Delete(&StrategyTab{}).Error
}

func FindStrategyByID(db storeMsql.DB, strategyID string) (ar *StrategyTab, err error) {
	res := &StrategyTab{}
	if err = db.GetDBConn().Model(&StrategyTab{}).Where(StrategyTab{StrategyID: strategyID}).First(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
