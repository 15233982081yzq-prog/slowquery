package mysql

import (
	storeMsql "smart-slowquery/pkg/store/mysql"

	"time"
)

type DailyNewFingerReportLog struct {
	TaskUUID    string    `gorm:"column:task_uuid;primaryKey;<-:create"`
	TaskName    string    `gorm:"column:task_name;primaryKey"`
	ProductLine string    `gorm:"column:product_line"`
	DBEnv       string    `gorm:"column:db_env"`
	Owners      string    `gorm:"column:owners"`
	Leaders     string    `gorm:"column:leaders"`
	NewFinger   int       `gorm:"new_finger"`
	NewSqlQuery int       `gorm:"new_sql_query"`
	ReportDay   time.Time `gorm:"column:report_day;primaryKey;<-:create"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
}

func FindDailyNewFingerReportLog(db storeMsql.DB, taskUUID, reportDay string) (*DailyNewFingerReportLog, error) {
	var result *DailyNewFingerReportLog
	err := db.Query(&result, "task_uuid = ? and report_day=?", taskUUID, reportDay)
	return result, err
}

func (dr *DailyNewFingerReportLog) TableName() string {
	return "slow_query_daily_new_finger_email_log"
}

func (dr *DailyNewFingerReportLog) save(db storeMsql.DB) error {
	return db.Save(dr)
}

func SaveDailyNewFingerReportLog(db storeMsql.DB, reportLog *DailyNewFingerReportLog) (err error) {
	return reportLog.save(db)
}
