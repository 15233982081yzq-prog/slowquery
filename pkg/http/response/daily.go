package response

type DBSlowQueryRankDailyVo struct {
	Rank    []*DBQueryTimeVo `json:"rank"`
	OrderBy string           `json:"order_by"`
	RankDay string           `json:"rank_day"`
	DBEnv   string           `json:"db_env"`
}

type DBQueryTimeVo struct {
	SerialNo    int      `json:"serial_no"`
	ClusterUUID string   `json:"cluster_uuid"`
	DBName      string   `json:"db_name"`
	OwnCMDB     string   `json:"own_cmdb"`
	ProductLine string   `json:"product_line"`
	Owners      []string `json:"owners"`
	//OwnerReportManager string
	Total      float64 `json:"total"`
	Count      int     `json:"count"`
	WeekOnWeek string  `json:"week_on_week"`
	RankDetail *Detail `json:"detail"`
}

type FingerSlowQueryRankDailyVo struct {
	Rank    []*FingerQueryTimeVo `json:"rank"`
	OrderBy string               `json:"order_by"`
	RankDay string               `json:"rank_day"`
	DBEnv   string               `json:"db_env"`
}

type FingerQueryTimeVo struct {
	SerialNo    int      `json:"serial_no"`
	FingerID    string   `json:"finger_id"`
	FingerSql   string   `json:"finger_sql"`
	ClusterUUID string   `json:"cluster_uuid"`
	DBName      string   `json:"db_name"`
	OwnCMDB     string   `json:"own_cmdb"`
	ProductLine string   `json:"product_line"`
	Owners      []string `json:"owners"`
	Total       float64  `json:"total"`
	Count       int      `json:"count"`
	WeekOnWeek  string   `json:"week_on_week"`
	RankDetail  *Detail  `json:"detail"`
}

type Detail struct {
	FingerID    string `json:"finger_id,omitempty"`
	DBEnv       string `json:"db_env,omitempty"`
	DBName      string `json:"db_name"`
	ClusterUUID string `json:"cluster_uuid"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
}
