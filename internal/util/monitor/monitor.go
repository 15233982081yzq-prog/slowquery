package monitor

import (
	sysMetrics "smart-slowquery/pkg/metrics/platform"
	storeReq "smart-slowquery/pkg/store/request"

	"smart-slowquery/pkg/log"

	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	client = &http.Client{
		Timeout: time.Second * 30,
	}

	SuccessStatus = "success"

	prometheusOpenAPI = "https://monitoring-dbproduct.infra.sz.shopee.io/vmselect/select/10625/prometheus/api/v1/query_range?query=%s&step=%d&start=%d&end=%d"
)

func InitMonitorPrometheusOpenApi(url string) {
	prometheusOpenAPI = url
}

func GetMetric(databases []string, metricName, metricType string, step int, startTime, endTime int64, appearType *storeReq.AppearType) (*MetricData, error) {
	var (
		start    time.Time
		queryUrl string
		err      error
		fData    prometheusFetchData
		req      *http.Request
		resp     *http.Response
	)
	start = time.Now()
	defer sysMetrics.CollectServiceMetrics("GetMetric", sysMetrics.GetStatus(err), time.Since(start))
	// 1. create request
	queryUrl = buildQuery(databases, metricName, metricType, step, startTime, endTime, appearType)
	if req, err = http.NewRequest(http.MethodGet, queryUrl, nil); err != nil {
		log.Errorf("GetMetric create request error %v url %s", err, queryUrl)
		return nil, err
	}

	// 4. make request
	if resp, err = client.Do(req); err != nil {
		log.Errorf("GetMetric do request error %v url %s", err, queryUrl)
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&fData); err != nil {
		log.Errorf("GetMetric decode error:%s", err.Error())
		return nil, err
	}

	if !fData.Success() {
		log.Errorf("GetMetric response error, error_type:%s, error_msg:%s", fData.ErrorType, fData.ErrorObj)
		return nil, fmt.Errorf("GetMetric get response error, url %s", queryUrl)
	}

	return &fData.Data, err
}

type prometheusFetchData struct {
	Status    string     `json:"status"`
	IsPartial bool       `json:"isPartial"`
	ErrorType string     `json:"errorType"`
	ErrorObj  string     `json:"error"`
	Warnings  []string   `json:"warnings"`
	Data      MetricData `json:"data"`
}

type MetricData struct {
	ResultType string `json:"resultType"`
	Result     []struct {
		Metric struct {
			Col          string `json:"col"`
			DatabaseName string `json:"database_name"`
		} `json:"metric"`
		Values [][]interface{} `json:"values"`
	} `json:"result"`
}

func (pfd *prometheusFetchData) Success() bool {
	return pfd.Status == SuccessStatus
}

func buildQuery(databases []string, metricName, metricType string, step int, startTime, endTime int64, appearType *storeReq.AppearType) string {
	if appearType.IsAll() {
		return fmt.Sprintf(prometheusOpenAPI, url.QueryEscape("sum by(database_name,col)")+fmt.Sprintf("(%s{database_name=~\"%s\",col=\"%s\"})", metricName, strings.Join(databases, "|"), metricType), step, startTime, endTime)
	}
	return fmt.Sprintf(prometheusOpenAPI, url.QueryEscape("sum by(database_name,col)")+fmt.Sprintf("(%s{database_name=~\"%s\",col=\"%s\",new_appeared=\"%d\"})", metricName, strings.Join(databases, "|"), metricType, appearType.GetSign()), step, startTime, endTime)
}
