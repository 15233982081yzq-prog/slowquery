package clickhouse

import (
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/alert"
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"

	"database/sql"
	"fmt"
	"strconv"
	"time"
)

func (cli *Client) getClientUsers(req *request.SlowQueryClientUsers) (clientUsers []string, err error) {
	tx := cli.db.Select("client_user,count(1) as count").Table(req.GetTableName()).Where("database_name = ?", req.DbName).Where("environment=?", req.DbEnv).Where("finger_id=?", req.FingerID)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}

	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	tx.Group("client_user").Order("count desc")

	var clientUsersStats []response.ClientUserStats

	if err = tx.Scan(&clientUsersStats).Error; err != nil {
		log.Errorf("ck client getClientUsers error:%s , sql:%s", err.Error(), tx.Statement.SQL.String())
	}

	for _, stats := range clientUsersStats {
		clientUsers = append(clientUsers, stats.ClientUser)
	}
	return
}

func (cli *Client) GetClientHostsStats(req *request.SlowQueryClientHostsStats) (stats *[]response.ClientHostStats, err error) {
	tx := cli.db.Select("client_host,count(1) as count").Table(req.GetTableName()).Where("database_name = ? and environment=?", req.DbName, req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}
	tx.Where("finger_id = ?", req.FingerID)

	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	tx.Group("client_host").Order("count desc")

	var list []response.ClientHostStats

	if err = tx.Scan(&list).Error; err != nil {
		log.Errorf("ck client getClientHostsStats error:%s", err.Error())
	}
	return &list, err
}

func (cli *Client) GetInstanceHosts(req *request.SlowQueryInstanceHosts) (instanceHosts []string, err error) {
	tx := cli.db.Select("instance_host").Table(req.GetTableName()).Where("database_name = ? and environment=?", req.DbName, req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}
	tx.Group("instance_host")

	if err = tx.Scan(&instanceHosts).Error; err != nil {
		log.Errorf("ck client GetInstanceHosts error:%s", err.Error())
	}
	return
}

func (cli *Client) GetFirstQueryStatement(req *request.SlowQueryStatement) (*response.QueryStatement, error) {
	tx := cli.db.Select("slow_sql,instance_host,instance_port,is_new_appeared,log_time").Table(req.GetTableName()).
		Where("database_name = ? and environment=?", req.DbName, req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}

	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	if !req.AppearType.IsAll() {
		tx.Where("is_new_appeared = ?", req.AppearType.GetSign())
	}
	tx.Where("finger_id = ?", req.FingerID)
	tx.Order("log_time asc")
	tx.Limit(1)

	var (
		resp response.QueryStatement
		err  error
	)
	if err = tx.Find(&resp).Error; err != nil {
		log.Errorf("ck client getQueryStatement error:%s", err.Error())
	}
	return &resp, err
}
func (cli *Client) GetQueryStatements(req *request.SlowQueryStatementWithOrderBy) (*[]response.QueryStatement, error) {
	tx := cli.db.Select("slow_sql,instance_host,instance_port,query_time,lock_time").Table(req.GetTableName()).
		Where("database_name = ? and environment=?", req.DbName, req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}

	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	tx.Where("finger_id", req.FingerID)
	tx.Order(fmt.Sprintf("%s desc", req.OrderBy))
	tx.Limit(req.Limit).Offset(req.Offset)

	var (
		stmts []response.QueryStatement
		err   error
	)
	if err = tx.Find(&stmts).Error; err != nil {
		log.Errorf("ck client getQueryStatements error:%s", err.Error())
	}
	return &stmts, err
}

func (cli *Client) GetQueryStatementsCount(req *request.SlowQueryStatementWithOrderBy) (int, error) {
	tx := cli.db.Select("count(1) as count").Table(req.GetTableName()).
		Where("database_name = ? and environment=?", req.DbName, req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}

	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	tx.Where("finger_id", req.FingerID)

	var (
		count int
		err   error
	)
	if err = tx.Find(&count).Error; err != nil {
		log.Errorf("ck client GetQueryStatementsCount error:%s", err.Error())
	}
	return count, err
}

func (cli *Client) GetSlowQueryDBStatistic(req *request.SlowQueryDBStatistic) (*[]response.SlowQueryDBStatistic, error) {
	tx := cli.db.Select("finger_id,finger_sql,cluster_uuid,database_name,sum(query_time) as total_query_time,count(1) as count,sum(lock_time) as total_lock_time, avg(query_time) as avg_query_time, any(role) as role, quantile(0.8)(query_time) as top80_time").
		Table(req.GetTableName()).Where("database_name in ?", req.DbNames)
	if len(req.ClusterUUids) > 0 {
		tx.Where("cluster_uuid in ?", req.ClusterUUids)
	}
	tx.Where("environment=?", req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}
	tx.Where("database_name not in ('information_schema','mysql', 'INFORMATION_SCHEMA')")

	tx.Group("finger_id,finger_sql,database_name,cluster_uuid")

	var (
		err  error
		list []response.SlowQueryDBStatistic
	)

	if err = tx.Scan(&list).Error; err != nil {
		log.Errorf("ck client GetSlowQueryDBStatistic error:%s", err.Error())
	}
	return &list, err
}

func (cli *Client) GetSlowQueryList(req *request.SlowQueryList) (*[]response.SlowQuery, error) {
	tx := cli.db.Select("finger_id,finger_sql,cluster_uuid,database_name,sum(query_time) as total_time,count(1) as count,avg(query_time) as avg_time,sum(is_new_appeared) as new_appear_flag").
		Table(req.GetTableName()).Where("database_name in ?", req.DbNames).Where("environment=?", req.DbEnv)

	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}
	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	if len(req.ClusterUUIDs) > 0 {
		tx.Where("cluster_uuid in ?", req.ClusterUUIDs)
	}
	tx.Group("finger_id,finger_sql,database_name,cluster_uuid")
	if req.AppearType.IsNewAppeared() {
		tx.Having("new_appear_flag > 0")
	}
	if req.AppearType.IsOriginal() {
		tx.Having("new_appear_flag = 0")
	}
	tx.Order(fmt.Sprintf("%s desc", req.OrderBy))
	tx.Limit(req.Limit).Offset(req.Offset)

	var (
		err  error
		list []response.SlowQuery
	)

	if err = tx.Scan(&list).Error; err != nil {
		log.Errorf("ck client GetSlowQueryList error:%s", err.Error())
	}
	return &list, err
}

func (cli *Client) GetLast7Days(req []string) ([]string, error) {
	list := make([]string, 0)
	if err := cli.db.Select("finger_id").Table("slow_query_log_last_7_days").Where("finger_id in ?", req).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (cli *Client) GetSlowQueryCount(req *request.SlowQueryCount) (count int, err error) {
	tx := cli.db.Select("count(1) as count,sum(is_new_appeared) as new_appear_flag").Table(req.GetTableName()).Where("database_name in ?", req.DbNames).Where("environment=?", req.DbEnv)
	if req.StartTime > 0 && req.EndTime > 0 {
		tx.Where("log_time >= ? and log_time <= ? ", time.Unix(req.StartTime, 0).Format(request.TimeFormat), time.Unix(req.EndTime, 0).Format(request.TimeFormat))
	}
	if len(req.Instances) > 0 {
		tx.Where("instance_host in ?", req.Instances)
	}
	if len(req.ClusterUUIDs) > 0 {
		tx.Where("cluster_uuid in ?", req.ClusterUUIDs)
	}
	tx.Group("finger_id")
	if req.AppearType.IsNewAppeared() {
		tx.Having("new_appear_flag > 0")
	}
	if req.AppearType.IsOriginal() {
		tx.Having("new_appear_flag = 0")
	}

	var rows *sql.Rows
	if rows, err = tx.Rows(); err != nil {
		log.Errorf("ck client GetSlowQueryCount error:%s", err.Error())
	}

	for rows.Next() {
		count++
	}
	return
}

// --------------------------------------------------------------------- SLowQuery 报表 ----------------------------------------------------------------------------//

func (cli *Client) DBSlowQueryRank(req *request.SlowQueryRank) (*[]response.CKDBQueryTime, error) {
	tx := cli.db.Select(fmt.Sprintf("database_name,cluster_uuid,ROUND(SUM(%s),2) AS total,count(1) as count", req.OrderBy)).Table(req.GetTableName()).
		Where("log_time >= ? and log_time<= ?", req.StartTime.Format(request.TimeFormat), req.EndTime.Format(request.TimeFormat))
	tx.Where("environment = ?", req.DBEnv)
	tx.Where("database_name not in ('information_schema','mysql')")
	tx.Group("database_name,cluster_uuid")
	tx.Order("total desc")
	tx.Limit(req.Top)

	var (
		err  error
		list []response.CKDBQueryTime
	)

	if err = tx.Scan(&list).Error; err != nil {
		log.Errorf("ck client DBSlowQueryRank error:%s", err.Error())
	}

	return &list, err
}

func (cli *Client) FingerSlowQueryRank(req *request.SlowQueryRank) (*[]response.CKFingerQueryTime, error) {
	tx := cli.db.Select(fmt.Sprintf("finger_id,finger_sql,cluster_uuid,database_name,ROUND(SUM(%s),2) AS total,count(1) as count", req.OrderBy)).Table(req.GetTableName()).
		Where("log_time >= ? and log_time<= ?", req.StartTime.Format(request.TimeFormat), req.EndTime.Format(request.TimeFormat))
	tx.Where("database_name not in ('information_schema','mysql')")
	tx.Where("environment=?", req.DBEnv)
	tx.Where("is_new_appeared > 0")
	tx.Group("finger_id,finger_sql,cluster_uuid,database_name")
	tx.Order("total desc")
	tx.Limit(req.Top)

	var (
		err  error
		list []response.CKFingerQueryTime
	)

	if err = tx.Scan(&list).Error; err != nil {
		log.Errorf("ck client FingerSlowQueryRank error:%s", err.Error())
	}

	return &list, err
}

func (cli *Client) NewFingerSlowQueryReportRecord(req *request.SlowQueryNewFinger) (list []*response.NewFingerReportRecord, err error) {
	tx := cli.db.Select(fmt.Sprintf("database_name,cluster_uuid,count(DISTINCT(finger_id)) as query_finger_count,count(finger_sql) as query_sql_count, min(log_time) as datetime, ROUND(SUM(%s),2) AS total", req.OrderBy)).Table(req.TableName()).
		Where("log_time >= ? and log_time<= ?", req.StartTime.Format(request.TimeFormat), req.EndTime.Format(request.TimeFormat))
	tx.Where("database_name not in ('information_schema','mysql')")
	tx = tx.Where("environment=?", req.DBEnv)
	if req.IsNewAppeared {
		tx = tx.Where("is_new_appeared > 0")
	}
	tx.Group("database_name,cluster_uuid")
	tx.Order("query_finger_count desc").Limit(req.Top)

	if err = tx.Scan(&list).Error; err != nil {
		log.Errorf("ck client NewFingerSlowQueryReportRecord error:%s", err.Error())
	}

	return list, err
}

func (cli *Client) GetAlertMessage(req *request.AlertMessageSearch) ([]*response.AlertMessageWithMuteOperator, error) {
	list := make([]*response.AlertMessageWithMuteOperator, 0)
	tx := cli.db.Table(fmt.Sprintf("%s as message", req.TableName())).
		Where("message.create_time >= ? and message.create_time <= ?", req.StartTime.Format(request.TimeFormat), req.EndTime.Format(request.TimeFormat))

	if req.IsMute {
		tx.Select("message.*, mute.creator as mute_operator, mute.end_time as mute_to").Joins("global left join alert_message_mute_all_rand as mute on message.monitor_alert_id = mute.monitor_alert_id")
		tx.Where("mute.end_time > ? and mute.status= ?", time.Now().Unix(), alert.MuteMutedStatus)
	} else {
		tx.Select("message.*")
		tx.Where("message.monitor_alert_id global not in (select mute.monitor_alert_id from alert_message_mute_all_rand as mute where mute.end_time > ? and mute.status= ?)", time.Now().Unix(), alert.MuteMutedStatus)
	}

	if len(req.CMDBs) > 0 {
		tx.Where("message.cmdb in ?", req.CMDBs)
	}

	if len(req.RuleName) > 0 {
		tx.Where("message.alert_rule_name = ?", req.RuleName)
	}

	if len(req.DataBaseName) > 0 {
		tx.Where("message.database_name like ?", "%"+req.DataBaseName+"%")
	}
	tx.Where("message.environment = ?", req.Env)
	if len(req.Severity) > 0 {
		tx.Where("message.severity = ?", req.Severity)
	}
	if len(req.TemplateName) > 0 {
		tx.Where("message.template_name = ?", req.TemplateName)
	}
	if len(req.Status) > 0 {
		tx.Where("message.alert_status = ?", req.Status)
	}

	tx.Order("message.update_time desc")
	if err := tx.Limit(req.Limit).Offset(req.Offset).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (cli *Client) GetAlertMessageCount(req *request.AlertMessageSearch) (int, error) {
	var count int64
	tx := cli.db.Table(fmt.Sprintf("%s as message", req.TableName()))

	if !req.StartTime.IsZero() && !req.EndTime.IsZero() {
		tx.Where("message.create_time >= ? and message.create_time <= ?", req.StartTime.Format(request.TimeFormat), req.EndTime.Format(request.TimeFormat))
	}
	if len(req.CMDBs) > 0 {
		tx.Where("message.cmdb in ?", req.CMDBs)
	}
	if len(req.DataBaseName) > 0 {
		tx.Where("message.database_name like ?", "%s"+req.DataBaseName+"%s")
	}
	tx.Where("message.environment = ?", req.Env)
	if len(req.Severity) > 0 {
		tx.Where("message.severity = ?", req.Severity)
	}

	if len(req.RuleName) > 0 {
		tx.Where("message.alert_rule_name = ?", req.RuleName)
	}

	if len(req.TemplateName) > 0 {
		tx.Where("message.template_name = ?", req.TemplateName)
	}
	if len(req.Status) > 0 {
		tx.Where("message.alert_status = ?", req.Status)
	}
	if req.IsMute {
		tx.Where("message.monitor_alert_id global in (select mute.monitor_alert_id from alert_message_mute_all_rand as mute where mute.end_time > ? and mute.status= ?)", time.Now().Unix(), alert.MuteMutedStatus)
	} else {
		tx.Where("message.monitor_alert_id global not in (select mute.monitor_alert_id from alert_message_mute_all_rand as mute where mute.end_time > ? and mute.status= ?)", time.Now().Unix(), alert.MuteMutedStatus)
	}
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (cli *Client) GetAlertMessageCountBySpecCond(req *request.AlertMessageCountSearch) (int, error) {
	var count int64
	tx := cli.db.Table(fmt.Sprintf("%s as message", req.TableName()))

	if !req.StartTime.IsZero() && !req.EndTime.IsZero() {
		tx.Where("message.create_time >= ? and message.create_time <= ?", req.StartTime.Format(request.TimeFormat), req.EndTime.Format(request.TimeFormat))
	}
	if len(req.CMDBs) > 0 {
		tx.Where("message.cmdb in ?", req.CMDBs)
	}
	tx.Where("message.environment = ?", req.Env)

	if req.IsMute {
		tx.Where("message.monitor_alert_id global in (select mute.monitor_alert_id from alert_message_mute_all_rand as mute where mute.end_time > ? and mute.status=?)", time.Now().Unix(), alert.MuteMutedStatus)
	}

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (cli *Client) GetUnMutedAndTTLStatusMessageList(req *request.AlertNoMutedMessage) ([]*response.AlertMessage, error) {
	list := make([]*response.AlertMessage, 0)
	tx := cli.db.Table(req.TableName())

	tx.Select("*").
		Joins("global left join alert_message_mute_all_rand on alert_message_log_all_rand.monitor_alert_id = alert_message_mute_all_rand.monitor_alert_id").
		Where("alert_message_log_all_rand.alert_status = ?", request.AlertPending).
		Where("alert_message_mute_all_rand.status=? or alert_message_mute_all_rand.status=?", alert.MuteUnMutedStatus, alert.MuteTTLStatus)

	if err := tx.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (cli *Client) GetAlertMessageByAlertID(alertID int64) (resp *response.AlertMessage, err error) {
	resp = &response.AlertMessage{}
	if err = cli.db.Model(&request.AlertMessage{}).Where("monitor_alert_id=?", strconv.Itoa(int(alertID))).First(resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (cli *Client) GetAlertMessageByMonitorRuleIdAndStatus(ruleID, status string) (resp []*response.AlertMessage, err error) {
	resp = make([]*response.AlertMessage, 0)
	if err = cli.db.Model(&request.AlertMessage{}).Where("alert_rule_uuid=? and alert_status=?", ruleID, status).Find(&resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (cli *Client) GetAlertMuteByAlertID(alertID string) (resp []*response.AlertMute, err error) {
	resp = make([]*response.AlertMute, 0)
	if err = cli.db.Model(&request.AlertMute{}).Where("monitor_alert_id = ? ", alertID).Find(&resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (cli *Client) GetAlertMessageByAlertIDs(alertIDs []string) (resp []*response.AlertMessage, err error) {
	resp = make([]*response.AlertMessage, 0)
	if err = cli.db.Model(&request.AlertMessage{}).Where("monitor_alert_id in ?", alertIDs).Find(&resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}
