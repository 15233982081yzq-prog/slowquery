package mysql

import (
	"smart-slowquery/pkg/log"
	storeMsql "smart-slowquery/pkg/store/mysql"

	"time"
)

type DBSlowQueryDailyRank struct {
	SerialNo    int       `gorm:"column:serial_no"`
	ClusterUUID string    `gorm:"column:cluster_uuid"`
	DBName      string    `gorm:"column:db_name;primaryKey;<-:create"`
	DBEnv       string    `gorm:"column:db_env"`
	RankOrder   string    `gorm:"column:rank_order"`
	RankScore   float64   `gorm:"column:rank_score"`
	SqlCount    int       `gorm:"column:sql_count"`
	WeekOnWeek  string    `gorm:"column:week_on_week"`
	RankDay     time.Time `gorm:"column:rank_day;primaryKey;<-:create"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
}

func (dr *DBSlowQueryDailyRank) TableName() string {
	return "slow_query_db_daily_rank"
}

func (dr *DBSlowQueryDailyRank) Save(db storeMsql.DB) error {
	return db.Save(dr)
}

func SaveDBDailyRank(db storeMsql.DB, rank []*DBSlowQueryDailyRank) (err error) {
	return db.Transaction(func(db storeMsql.DB) error {
		for i := range rank {
			daily := rank[i]
			if err = daily.Save(db); err != nil {
				log.Errorf("CreateDBDailyRank message:%v ,error:%s", daily, err.Error())
				return err
			}
		}
		return nil
	})
}

func FindDBDailyRank(db storeMsql.DB, order, time, dbEnv string, top int) ([]*DBSlowQueryDailyRank, error) {
	var sqls []*DBSlowQueryDailyRank
	err := db.Query(&sqls, "rank_day=? and rank_order= ? and db_env= ? order by serial_no asc limit ?", time, order, dbEnv, top)
	return sqls, err
}

type FingerSlowQueryDailyRank struct {
	SerialNo    int       `gorm:"column:serial_no"`
	FingerID    string    `gorm:"finger_id;primaryKey;<-:create"`
	FingerSql   string    `gorm:"finger_sql"`
	ClusterUUID string    `gorm:"column:cluster_uuid"`
	DBName      string    `gorm:"column:db_name;primaryKey;<-:create"`
	DBEnv       string    `gorm:"column:db_env"`
	RankOrder   string    `gorm:"column:rank_order"`
	RankScore   float64   `gorm:"column:rank_score"`
	SqlCount    int       `gorm:"column:sql_count"`
	WeekOnWeek  string    `gorm:"column:week_on_week"`
	RankDay     time.Time `gorm:"column:rank_day;primaryKey;<-:create"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
}

func (dr *FingerSlowQueryDailyRank) TableName() string {
	return "slow_query_finger_daily_rank"
}

func (dr *FingerSlowQueryDailyRank) Save(db storeMsql.DB) error {
	return db.Save(dr)
}

func FindFingerDailyRank(db storeMsql.DB, order, time, dbEnv string, top int) ([]*FingerSlowQueryDailyRank, error) {
	var sqls []*FingerSlowQueryDailyRank
	err := db.Query(&sqls, "rank_day=? and rank_order= ? and db_env = ? order by serial_no asc limit ?", time, order, dbEnv, top)
	return sqls, err
}

func SaveFingerDailyRank(db storeMsql.DB, rank []*FingerSlowQueryDailyRank) (err error) {
	return db.Transaction(func(db storeMsql.DB) error {
		for i := range rank {
			daily := rank[i]
			if err = daily.Save(db); err != nil {
				log.Errorf("CreateFingerDailyRank message:%v ,error:%s", daily, err.Error())
				return err
			}
		}
		return nil
	})
}
