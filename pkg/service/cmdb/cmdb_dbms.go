package cmdb

import (
	"smart-slowquery/pkg/log"
	"time"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	DBTypeMySQL = "mysql"
	DBTypeTiDB  = "tidb"
)

// errors
var (
	client = &http.Client{
		Timeout: time.Second * 30,
	}

	ErrNoDBUuid      = errors.New("no corresponding database uuid, please contact DBA")
	ErrDBRetired     = errors.New("the db is retired, please contact DBA")
	ErrWrongDBEnv    = errors.New("database env is not right")
	ErrWrongDBType   = errors.New("database type is not right")
	ErrNoSlaveInst   = errors.New("the db has no slave instance")
	DBAMetaTicketUrl = "https://space.shopee.io/dba/dbmeta/ticket/table_change/shopee"
)

func GetDataBaseDetail(dbName, env, token, spaceHost string) (*GetDatabaseDetailResponse, error) {
	queryUrl := spaceHost + DBAMetaGetDatabase
	// 1. create request
	req, err := http.NewRequest(http.MethodGet, queryUrl, nil)

	if err != nil {
		log.Errorf("create request error %v url %s", err, queryUrl)
		return nil, err
	}

	// 2. concat query string
	queryParams := make(url.Values)
	queryParams.Add("database_name", dbName)
	queryParams.Add("environment", env)
	queryParams.Add("database_type", DBTypeMySQL)
	req.URL.RawQuery = queryParams.Encode()

	// 3. add token header
	req.Header.Add("AuthToken", token)

	// 4. make request
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("make request error %v url %s", err, queryUrl)
		return nil, err
	}
	defer resp.Body.Close()

	var databaseDetailResp GetDatabaseDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&databaseDetailResp); err != nil {
		log.Errorf("GetServiceTree decode error:%s", err.Error())
		return nil, err
	}

	if databaseDetailResp.ErrorCode != 0 {
		log.Errorf("response error, error_message:%s, error_code:%d", databaseDetailResp.Error, databaseDetailResp.ErrorCode)
		return nil, fmt.Errorf("get response error, url %s,error_message:%s,error_code:%d", queryUrl, databaseDetailResp.Error, databaseDetailResp.ErrorCode)
	}

	return &databaseDetailResp, nil
}

func GetServiceTree(auth, fuzzySearchName, spaceHost string) (*GetServiceTreeResponse, error) {
	// 1. create request
	var queryUrl string
	queryUrl = spaceHost + CMDBServiceGetTree
	req, err := http.NewRequest(http.MethodGet, queryUrl, nil)

	if err != nil {
		log.Errorf("create request error %v url %s", err, queryUrl)
		return nil, err
	}

	// 2. concat query string
	queryParams := make(url.Values)
	queryParams.Add("name", fuzzySearchName)
	req.URL.RawQuery = queryParams.Encode()

	// 3. add token header
	req.Header.Add("Authorization", auth)

	// 4. make request
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("make request error %v url %s", err, queryUrl)
		return nil, err
	}
	defer resp.Body.Close()

	var getServiceTreeResponse GetServiceTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&getServiceTreeResponse); err != nil {
		log.Errorf("GetServiceTree decode error:%s", err.Error())
		return nil, err
	}

	if !getServiceTreeResponse.Success {
		log.Errorf("response error, error_code:%d, error_msg:%t", getServiceTreeResponse.BusinessCode, getServiceTreeResponse.Success)
		return nil, fmt.Errorf("get response error, url %s", queryUrl)
	}

	return &getServiceTreeResponse, nil
}
