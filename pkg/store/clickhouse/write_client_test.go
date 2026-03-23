package clickhouse

import (
	"fmt"
	"smart-slowquery/pkg/store/request"
	"time"
)

// fake 100 data into db
func initDatabaseData(ck *Client) (data []*request.SlowQueryLog, err error) {
	tmpData := fakeData()
	return tmpData, ck.batchPut(tmpData)
}

func cleanDatabaseData(ck *CKStore, data []*request.SlowQueryLog) error {
	for i := 0; i < len(data); i++ {
		if err := ck.client.db.Where("finger_id=?", data[i].FingerID).Delete(data[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func fakeData() []*request.SlowQueryLog {
	tmpData := make([]*request.SlowQueryLog, 100)
	for i := 0; i < len(tmpData); i++ {
		temNo := i % 10
		tmpData[i] = &request.SlowQueryLog{
			FingerID:     fmt.Sprintf("finger_id_%d", temNo),
			FingerSql:    fmt.Sprintf("finger_sql_%d", temNo),
			Query:        fmt.Sprintf("query_%d", temNo),
			Hint:         fmt.Sprintf("hint_%d", temNo),
			ClusterUUID:  fmt.Sprintf("uuid_%d", temNo),
			Host:         fmt.Sprintf("host_%d", temNo),
			Port:         3306,
			DBName:       fmt.Sprintf("db_%d", temNo),
			DBEnv:        fmt.Sprintf("env_%d", temNo),
			QueryTime:    1,
			LockTime:     1,
			ExaminedRows: 0,
			NumRows:      0,
			AffectRows:   0,
			BytesSent:    100.,
			ClientIP:     fmt.Sprintf("127.0.0.%d", temNo),
			ConnectionID: 0,
			DefaultUser:  fmt.Sprintf("user_%d", temNo),
			User:         fmt.Sprintf("user_%d", temNo),
			LogTime:      time.Now().AddDate(0, 0, -i).Format("2006-01-02 15:04:05"),
			CreateTime:   time.Now().AddDate(0, 0, -i).Format("2006-01-02 15:04:05"),
			Killed:       0,
			LastErrno:    0,
		}
	}
	return tmpData
}
