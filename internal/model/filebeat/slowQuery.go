package filebeat

import timeUtil "smart-slowquery/internal/util/time"

type SlowQuery struct {
	TimeStamp          string        `json:"@timestamp"`
	MData              *MetaData     `json:"@metadata"`
	Message            string        `json:"message"`
	Fields             *Field        `json:"fields"`
	SlowLog            *MysqlSlowLog `json:"-"`
	IsNewAppeared      int           `json:"-"`
	HasOpenIndexSwitch int           `json:"-"`
	IsHitIndex         int           `json:"-"`
	L1L2               string        `json:"-"`
	Team               string        `json:"-"`
	Role               string        `json:"-"`
}

type MetaData struct {
	Beat    string `json:"beat"`
	Typ     string `json:"type"`
	Version string `json:"version"`
}

type Field struct {
	ClusterUUID  string `json:"cluster_uuid"`
	ClusterType  string `json:"cluster_type"`
	InstanceHost string `json:"host"`
	InstancePort int32  `json:"port"`
	Env          string `json:"environment"`
}

type MysqlSlowLog struct {
	ConnectionID int     `mapstructure:"connectionId"`
	CurrentDB    string  `mapstructure:"currentDB"`
	AffectRows   int     `mapstructure:"affectedRows"`
	ExaminedRows int     `mapstructure:"examinedRows"`
	NumRows      int     `mapstructure:"numRows"`
	Killed       int     `mapstructure:"killed"`
	LastErrno    int     `mapstructure:"lastErrno"`
	BytesSent    int     `mapstructure:"bytesSent"`
	TimeStamp    int64   `mapstructure:"timestamp"`
	LockTime     float64 `mapstructure:"lockTime"`
	QueryTime    float64 `mapstructure:"queryTime"`
	Query        string  `mapstructure:"query"`
	Hint         string  `mapstructure:"hint"`
	ClientIP     string  `mapstructure:"ClientIP"`
	DefaultUser  string  `mapstructure:"defaultUser"`
	User         string  `mapstructure:"user"`
	FingerID     string  `mapstructure:"fingerId"`
	FingerSql    string  `mapstructure:"finger_sql"`
	ClusterType  string
}

func (msl *SlowQuery) Valid() bool {

	slowLog := msl.SlowLog
	if slowLog == nil {
		return false
	}

	//len(currentDB) <= 0 的判断说明slow query的命令查询/操作的是系统的内置数据库，TODO：需要对齐需求
	if len(slowLog.CurrentDB) <= 0 || len(slowLog.Query) <= 0 || len(slowLog.FingerID) <= 0 || len(slowLog.FingerSql) <= 0 {
		return false
	}

	if slowLog.TimeStamp == 0 || slowLog.QueryTime == 0 {
		return false
	}

	if len(msl.Fields.Env) == 0 || len(msl.Fields.InstanceHost) == 0 || len(msl.Fields.ClusterUUID) == 0 || msl.Fields.InstancePort <= 0 {
		return false
	}

	if slowLog.TimeStamp < timeUtil.GetUnixTimeStamp(-3) {
		return false
	}

	return true
}
