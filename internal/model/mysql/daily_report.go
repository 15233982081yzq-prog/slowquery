package mysql

import (
	storeMsql "smart-slowquery/pkg/store/mysql"

	"time"
)

type DailyReportLog struct {
	TaskUUID    string    `gorm:"column:task_uuid;primaryKey;<-:create"`
	TaskName    string    `gorm:"column:task_name;primaryKey"`
	ProductLine string    `gorm:"column:product_line"`
	DBEnv       string    `gorm:"column:db_env"`
	Owners      string    `gorm:"column:Owners"`
	Leaders     string    `gorm:"column:leaders"`
	ReportDay   time.Time `gorm:"column:report_day;primaryKey;<-:create"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
}

func (dr *DailyReportLog) TableName() string {
	return "slow_query_daily_rank_email_log"
}

func (dr *DailyReportLog) save(db storeMsql.DB) error {
	return db.Save(dr)
}

func SaveDailyReportLog(db storeMsql.DB, reportLog *DailyReportLog) (err error) {
	return reportLog.save(db)
}

func FindDailyReportLog(db storeMsql.DB, taskUUID, reportDay string) (*DailyReportLog, error) {
	var result *DailyReportLog
	err := db.Query(&result, "task_uuid = ? and report_day=?", taskUUID, reportDay)
	return result, err
}
