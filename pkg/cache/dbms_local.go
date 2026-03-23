package cache

import "sync"

type DBMSInfo struct {
	DBName  string
	L1L2    string
	Team    string
	RoleMap map[string]string // key:ip value:role[master,slave]
}

var dbmsData sync.Map

func init() {
	dbmsData = sync.Map{}
}

func StoreData(key, value any) {
	dbmsData.Store(key, value)
}

func FeetchData(key any) (value DBMSInfo, exist bool) {
	d, ok := dbmsData.Load(key)
	if ok {
		return d.(DBMSInfo), true
	}
	return DBMSInfo{}, false
}
