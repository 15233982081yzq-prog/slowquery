package request

import "fmt"

type QueryDBStatistic struct {
	DBEnv     string            `json:"db_env"`
	DBs       []*QueryDBMeta    `json:"dbs"`
	StartTime int64             `json:"start_time"`
	EndTime   int64             `json:"end_time"`
	mp        map[string]string `json:"-"`
}

type QueryDBMeta struct {
	DBName      string `json:"db_name"`
	ClusterUUid string `json:"cluster_uuid"`
}

func (qds *QueryDBStatistic) GetDBNames() (names []string) {
	uniq := make(map[string]bool)
	for _, db := range qds.DBs {
		if _, ok := uniq[db.DBName]; !ok {
			names = append(names, db.DBName)
			uniq[db.DBName] = true
		}
	}
	return names
}

func (qds *QueryDBStatistic) GetClusterUUids() (clusterUUids []string) {
	uniq := make(map[string]bool)
	for _, db := range qds.DBs {
		if _, ok := uniq[db.ClusterUUid]; !ok {
			clusterUUids = append(clusterUUids, db.ClusterUUid)
			uniq[db.ClusterUUid] = true
		}
	}
	return clusterUUids
}

func (qds *QueryDBStatistic) ValidMapping(clusterUUid, dbName string) bool {
	qds.buildDBClusterMapping()
	_, ok := qds.mp[buildClusterDBKey(clusterUUid, dbName)]
	return ok
}

func (qds *QueryDBStatistic) buildDBClusterMapping() {
	mp := make(map[string]string, len(qds.DBs))
	for _, db := range qds.DBs {
		mp[buildClusterDBKey(db.ClusterUUid, db.DBName)] = db.ClusterUUid
	}
	qds.mp = mp
}

func buildClusterDBKey(clusterUUid, dbName string) string {
	return fmt.Sprintf("%s_%s", clusterUUid, dbName)
}
