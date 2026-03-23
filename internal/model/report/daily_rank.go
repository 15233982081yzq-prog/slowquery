package report

import "time"

type DBSlowQueryRankDaily struct {
	Rank        []*DBQueryTime `json:"rank"`
	OrderBy     string         `json:"order_by"`
	RankDay     string         `json:"rank_day"`
	RankDayTime time.Time      `json:"-"`
	DBEnv       string         `json:"db_env"`
}

type DBQueryTime struct {
	SerialNo    int      `json:"serial_no"`
	ClusterUUID string   `json:"cluster_uuid"`
	DBName      string   `json:"db_name"`
	OwnCMDB     string   `json:"own_cmdb"`
	ProductLine string   `json:"product_line"`
	Owners      []string `json:"owners"`
	Leaders     []string `json:"leaders"`
	//OwnerReportManager string
	Total      float64 `json:"total"`
	Count      int     `json:"count"`
	WeekOnWeek string  `json:"week_on_week"`
	DetailLink string  `json:"link"`
}

func (dbRank *DBSlowQueryRankDaily) GetRankDBNames() (names []string) {
	for _, one := range dbRank.Rank {
		names = append(names, one.DBName)
	}
	return names
}

type FingerSlowQueryRankDaily struct {
	Rank        []*FingerQueryTime `json:"rank"`
	OrderBy     string             `json:"order_by"`
	RankDay     string             `json:"rank_day"`
	RankDayTime time.Time          `json:"-"`
	DBEnv       string             `json:"db_env"`
}

type NewFingerDailyReport struct {
	NewFingerInfos []*NewFingerInfo `json:"new_finger_infos"`
	OrderBy        string           `json:"order_by"`
	Top            int              `json:"top"`
	ReportDay      string           `json:"rank_day"`
	ReportDayTime  time.Time        `json:"-"`
	DBEnv          string           `json:"db_env"`
}

func (dbRank *NewFingerDailyReport) GetRankDBNames() (names []string) {
	for _, one := range dbRank.NewFingerInfos {
		names = append(names, one.DBName)
	}
	return names
}

func (dbRank *NewFingerDailyReport) NewFingerCount() (count int) {
	for _, one := range dbRank.NewFingerInfos {
		count += one.NewFinger
	}
	return count
}

func (dbRank *NewFingerDailyReport) NewSqlQueryCount() (count int) {
	for _, one := range dbRank.NewFingerInfos {
		count += one.NewSqlQuery
	}
	return count
}

type FingerQueryTime struct {
	SerialNo    int      `json:"serial_no"`
	FingerID    string   `json:"finger_id"`
	FingerSql   string   `json:"finger_sql"`
	ClusterUUID string   `json:"cluster_uuid"`
	DBName      string   `json:"db_name"`
	OwnCMDB     string   `json:"own_cmdb"`
	ProductLine string   `json:"product_line"`
	Owners      []string `json:"owners"`
	Leaders     []string `json:"leaders"`
	Total       float64  `json:"total"`
	Count       int      `json:"count"`
	WeekOnWeek  string   `json:"week_on_week"`
	DetailLink  string   `json:"link"`
}

type NewFingerInfo struct {
	SerialNo    int      `json:"serial_no"`
	DBName      string   `json:"db_name"`
	ClusterUUID string   `json:"cluster_uuid"`
	OwnCMDB     string   `json:"own_cmdb"`
	ProductLine string   `json:"product_line"`
	Owners      []string `json:"owners"`
	Leaders     []string `json:"leaders"`
	NewFinger   int      `json:"new_finger"`
	NewSqlQuery int      `json:"new_sql_query"`
	Datetime    string   `json:"datetime"`
	DetailLink  string   `json:"detail_link"`
}

type NewFingerInfos []*NewFingerInfo

func (n NewFingerInfos) NewFingerCount() (count int) {
	for _, one := range n {
		count += one.NewFinger
	}
	return count
}

func (n NewFingerInfos) NewSqlQueryCount() (count int) {
	for _, one := range n {
		count += one.NewSqlQuery
	}
	return count
}

func (dbRank *FingerSlowQueryRankDaily) GetRankDBNames() (names []string) {
	for _, one := range dbRank.Rank {
		names = append(names, one.DBName)
	}
	return names
}
