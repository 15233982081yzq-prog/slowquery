package email

import (
	modelReport "smart-slowquery/internal/model/report"
	emailUtil "smart-slowquery/internal/util/email"
	timeUtil "smart-slowquery/internal/util/time"
	storeMsql "smart-slowquery/pkg/store/mysql"

	"smart-slowquery/conf"
	"smart-slowquery/internal/model/mysql"
	"smart-slowquery/pkg/log"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	dbTop10        = "slow_query_db_top_10"
	dbBizTop10     = "slow_query_db_biz_top_10"
	fingerTop10    = "slow_query_finger_top_10"
	fingerBizTop10 = "slow_query_finger_biz_top_10"
	newFinger      = "slow_query_new_finger_report"
	dbProductLine  = "db_product"
)

type RankReportService struct {
	config  *conf.ReportEmailConfig
	mysqlDB storeMsql.DB
}

func NewRankReportService(config *conf.ReportEmailConfig, mysqlDB storeMsql.DB) (*RankReportService, error) {
	if config == nil || mysqlDB == nil {
		return nil, fmt.Errorf("config or mysqlDB is empty")
	}
	return &RankReportService{
		config:  config,
		mysqlDB: mysqlDB,
	}, nil
}

func (srv *RankReportService) SendDBRankReport(dailyRank *modelReport.DBSlowQueryRankDaily) error {
	var err error

	log.Infof("RankReportService SendDBRankReport start ...")
	// check if need send email
	if !srv.config.RankSwitch {
		log.Errorf("RankReportService SendDBRankReport switch is close")
		return fmt.Errorf("report email switch is close")
	}

	if dailyRank == nil || len(dailyRank.Rank) == 0 {
		log.Errorf("RankReportService SendDBRankReport rank is empty")
		return fmt.Errorf("db rank is empty")
	}

	// add DetailUrl
	srv.buildDBDetailUrl(dailyRank)

	log.Infof("RankReportService SendDBRankReport rank size:%d", len(dailyRank.Rank))

	if err = srv.sendDBRankToInternal(dailyRank); err != nil {
		log.Errorf("RankReportService SendDBRankReport sendDBRankToInternal error:%s", err.Error())
		return err
	}
	return srv.sendDBRankToBizUsers(dailyRank)
}

func (srv *RankReportService) sendDBRankToInternal(dailyRank *modelReport.DBSlowQueryRankDaily) error {
	var (
		err                   error
		mailBody, mailSubject string
		ccs, recipients       []string
		reportDay             = dailyRank.RankDay
		taskUUID              = buildTaskUUID(dbTop10, dbProductLine, reportDay, dailyRank.GetRankDBNames())
	)

	if srv.isSend(reportDay, taskUUID) {
		log.Warningf("RankReportService sendDBRankToInternal taskUUID:%s has finished", taskUUID)
		return nil
	}

	// organize emil content
	mailBody, err = emailUtil.BuildEmailContent(srv.config.DBReportTemplatePath, dailyRank.Rank)
	if err != nil {
		log.Errorf("RankReportService sendDBRankToInternal BuildEmailContent error:%s", err.Error())
		return err
	}

	mailSubject = fmt.Sprintf("[SLOW QUERY] [%s] DB Slow Log Rank TOP 10 for %s", dailyRank.DBEnv, reportDay)
	recipients = strings.Split(srv.config.OwnerEmails, ",")

	// send email
	if err = emailUtil.SendMail(srv.config.SMTPAddress, srv.config.MailSender, mailSubject, mailBody, recipients, ccs, srv.config.User, srv.config.Pwd); err != nil {
		log.Errorf("RankReportService sendDBRankToInternal SendMail error:%s", err.Error())
		return err
	}

	return srv.saveReportLog(taskUUID, dbTop10, dbProductLine, recipients, recipients, dailyRank.RankDayTime)
}

func (srv *RankReportService) sendDBRankToBizUsers(dailyRank *modelReport.DBSlowQueryRankDaily) error {
	var reportDay = dailyRank.RankDay

	L2Mapping := make(map[string][]*modelReport.DBQueryTime, 0)

	for idx := range dailyRank.Rank {
		subList := L2Mapping[dailyRank.Rank[idx].ProductLine]
		subList = append(subList, dailyRank.Rank[idx])
		L2Mapping[dailyRank.Rank[idx].ProductLine] = subList
	}

	for pl, subRank := range L2Mapping {
		var recipients, ccs []string

		taskUUID := buildTaskUUID(dbBizTop10, pl, reportDay, getDBNames(subRank))

		if srv.isSend(reportDay, taskUUID) {
			log.Warningf("RankReportService sendDBRankToBizUsers taskUUID:%s has finished", taskUUID)
			continue
		}

		// organize emil content
		mailBody, err := emailUtil.BuildEmailContent(srv.config.DBReportTemplatePath, subRank)
		if err != nil {
			log.Errorf("RankReportService sendDBRankToBizUsers BuildEmailContent error:%s", err.Error())
			continue
		}
		mailSubject := fmt.Sprintf("[SLOW QUERY] [%s] DB Slow Log Rank TOP 10 for %s", dailyRank.DBEnv, reportDay)
		ccs = strings.Split(srv.config.OwnerEmails, ",")
		//recipients = append(recipients, subRank[0].Owners...) //TODO 上线前开启封印
		recipients = []string{"jian.bian@shopee.com", "hanwen.liu@shopee.com"}
		// send email
		if err = emailUtil.SendMail(srv.config.SMTPAddress, srv.config.MailSender, mailSubject, mailBody, recipients, ccs, srv.config.User, srv.config.Pwd); err != nil {
			log.Errorf("RankReportService sendDBRankToBizUsers SendMail error:%s", err.Error())
			continue
		}

		if err = srv.saveReportLog(taskUUID, dbBizTop10, pl, subRank[0].Owners, subRank[0].Leaders, dailyRank.RankDayTime); err != nil {
			log.Errorf("RankReportService sendDBRankToBizUsers saveReportLog error:%s", err.Error())
			continue
		}
	}

	return nil
}

func (srv *RankReportService) SendFingerRankReport(dailyRank *modelReport.FingerSlowQueryRankDaily) error {
	var err error

	// check if need send email
	if !srv.config.RankSwitch {
		log.Errorf("RankReportService SendFingerRankReport switch is close")
		return fmt.Errorf("report email switch is close")
	}

	if dailyRank == nil || len(dailyRank.Rank) == 0 {
		log.Errorf("RankReportService SendFingerRankReport rank is empty")
		return fmt.Errorf("db rank is empty")
	}

	// add DetailUrl
	srv.buildFingerDetailUrl(dailyRank)

	if err = srv.sendFingerRankToInternal(dailyRank); err != nil {
		log.Errorf("RankReportService SendFingerRankReport sendFingerRankToInternal error:%s", err.Error())
		return err
	}

	return srv.sendFingerRankToBizUsers(dailyRank)
}

func (srv *RankReportService) SendNewFingerRankReport(dailyRankReportRecord *modelReport.NewFingerDailyReport) error {

	// check if need send email
	if !srv.config.NewFingerSwitch {
		log.Errorf("RankReportService SendNewFingerRankReport switch is close")
		return fmt.Errorf("report email new finger switch is close")
	}

	if dailyRankReportRecord == nil || len(dailyRankReportRecord.NewFingerInfos) == 0 {
		log.Errorf("RankReportService SendNewFingerRankReport rank is empty")
		return fmt.Errorf("db record is empty")
	}

	// add DetailUrl
	srv.buildNewFingerDetailUrl(dailyRankReportRecord)

	return srv.sendNewFingerRankToInternal(dailyRankReportRecord)
}

func (srv *RankReportService) sendFingerRankToInternal(dailyRank *modelReport.FingerSlowQueryRankDaily) error {
	var (
		err                   error
		mailBody, mailSubject string
		ccs, recipients       []string
		reportDay             = dailyRank.RankDay
		taskUUID              = buildTaskUUID(fingerTop10, dbProductLine, reportDay, dailyRank.GetRankDBNames())
	)

	if srv.isSend(reportDay, taskUUID) {
		log.Warningf("RankReportService sendFingerRankToInternal taskUUID:%s has finished", taskUUID)
		return nil
	}

	// organize emil content
	mailBody, err = emailUtil.BuildEmailContent(srv.config.FingerReportTemplatePath, dailyRank.Rank)
	if err != nil {
		log.Errorf("RankReportService sendFingerRankToInternal BuildEmailContent error:%s", err.Error())
		return err
	}

	mailSubject = fmt.Sprintf("[SLOW QUERY] [%s] Finger Slow Log Rank TOP 10 for %s", dailyRank.DBEnv, reportDay)
	recipients = strings.Split(srv.config.OwnerEmails, ",")
	// send email
	if err = emailUtil.SendMail(srv.config.SMTPAddress, srv.config.MailSender, mailSubject, mailBody, recipients, ccs, srv.config.User, srv.config.Pwd); err != nil {
		log.Errorf("RankReportService sendFingerRankToInternal SendMail error:%s", err.Error())
		return err
	}

	return srv.saveReportLog(taskUUID, fingerTop10, dbProductLine, recipients, recipients, dailyRank.RankDayTime)
}

func (srv *RankReportService) sendNewFingerRankToInternal(dailyReport *modelReport.NewFingerDailyReport) error {
	var (
		err                   error
		mailBody, mailSubject string
		ccs, recipients       []string
		reportDay             = dailyReport.ReportDay
		taskUUID              = buildTaskUUID(newFinger, dbProductLine, reportDay, []string{})
	)

	if srv.newFingerReportIsSend(reportDay, taskUUID) {
		log.Warningf("RankReportService sendNewFingerRankToInternal taskUUID:%s has finished", taskUUID)
		return nil
	}

	for idx, info := range dailyReport.NewFingerInfos {
		info.SerialNo = idx + 1
	}

	// organize emil content
	mailBody, err = emailUtil.BuildEmailContent(srv.config.NewFingerReportTemplatePath, dailyReport.NewFingerInfos)
	if err != nil {
		log.Errorf("RankReportService sendNewFingerRankToInternal BuildEmailContent error:%s", err.Error())
		return err
	}

	mailSubject = fmt.Sprintf("[SLOW QUERY] [%s] Slow Log New Finger Appeared for %s", dailyReport.DBEnv, reportDay)
	recipients = strings.Split(srv.config.OwnerEmails, ",")
	// send email
	if err = emailUtil.SendMail(srv.config.SMTPAddress, srv.config.MailSender, mailSubject, mailBody, recipients, ccs, srv.config.User, srv.config.Pwd); err != nil {
		log.Errorf("RankReportService sendFingerRankToInternal SendMail error:%s", err.Error())
		return err
	}

	return srv.saveNewFingerReportLog(taskUUID, newFinger, dbProductLine, recipients, recipients, dailyReport.NewFingerCount(), dailyReport.NewSqlQueryCount(), dailyReport.ReportDayTime)
}

func (srv *RankReportService) sendFingerRankToBizUsers(dailyRank *modelReport.FingerSlowQueryRankDaily) error {
	var (
		reportDay = dailyRank.RankDay
		L2Mapping = make(map[string][]*modelReport.FingerQueryTime, 0)
	)

	for idx := range dailyRank.Rank {
		subList := L2Mapping[dailyRank.Rank[idx].ProductLine]
		subList = append(subList, dailyRank.Rank[idx])
		L2Mapping[dailyRank.Rank[idx].ProductLine] = subList
	}

	for pl, subRank := range L2Mapping {
		var recipients, ccs []string

		taskUUID := buildTaskUUID(fingerBizTop10, pl, reportDay, getFingerNames(subRank))

		if srv.isSend(reportDay, taskUUID) {
			log.Warningf("RankReportService sendFingerRankToBizUsers taskUUID:%s has finished", taskUUID)
			continue
		}

		// organize emil content
		mailBody, err := emailUtil.BuildEmailContent(srv.config.FingerReportTemplatePath, subRank)
		if err != nil {
			log.Errorf("RankReportService sendFingerRankToBizUsers BuildEmailContent error:%s", err.Error())
			return err
		}

		mailSubject := fmt.Sprintf("[SLOW QUERY] [%s] Finger Slow Log Rank TOP 10 for %s", dailyRank.DBEnv, reportDay)
		ccs = strings.Split(srv.config.OwnerEmails, ",")
		//recipients = append(recipients, subRank[0].Owners...) //TODO 上线前开启封印
		recipients = []string{"jian.bian@shopee.com", "hanwen.liu@shopee.com"}

		// send email
		if err = emailUtil.SendMail(srv.config.SMTPAddress, srv.config.MailSender, mailSubject, mailBody, recipients, ccs, srv.config.User, srv.config.Pwd); err != nil {
			log.Errorf("RankReportService sendFingerRankToBizUsers SendMail error:%s", err.Error())
			continue
		}
		if err = srv.saveReportLog(taskUUID, fingerBizTop10, subRank[0].ProductLine, subRank[0].Owners, subRank[0].Leaders, dailyRank.RankDayTime); err != nil {
			log.Errorf("RankReportService sendFingerRankToBizUsers saveReportLog error:%s", err.Error())
			continue
		}
	}
	return nil
}

func (srv *RankReportService) sendNewFingerRankToBizUsers(dailyReport *modelReport.NewFingerDailyReport) error {
	var (
		reportDay = dailyReport.ReportDay
		L2Mapping = make(map[string]modelReport.NewFingerInfos)
	)

	for idx := range dailyReport.NewFingerInfos {
		subList := L2Mapping[dailyReport.NewFingerInfos[idx].ProductLine]
		subList = append(subList, dailyReport.NewFingerInfos[idx])
		L2Mapping[dailyReport.NewFingerInfos[idx].ProductLine] = subList
	}

	for pl, subRecord := range L2Mapping {
		var recipients, ccs []string

		// 重新调整序号
		for idx, record := range subRecord {
			record.SerialNo = idx + 1
		}

		taskUUID := buildTaskUUID(newFinger, pl, reportDay, getNewFingerNames(subRecord))

		if srv.newFingerReportIsSend(reportDay, taskUUID) {
			log.Warningf("RankReportService sendNewFingerRankToBizUsers taskUUID:%s has finished", taskUUID)
			continue
		}

		// organize emil content
		mailBody, err := emailUtil.BuildEmailContent(srv.config.NewFingerReportTemplatePath, subRecord)
		if err != nil {
			log.Errorf("RankReportService sendNewFingerRankToBizUsers BuildEmailContent error:%s", err.Error())
			return err
		}

		mailSubject := fmt.Sprintf("[SLOW QUERY] [%s] Slow Log New Finger Appear for %s", dailyReport.DBEnv, reportDay)
		ccs = strings.Split(srv.config.OwnerEmails, ",")
		//recipients = append(recipients, subRank[0].Owners...) //TODO 上线前开启封印
		recipients = []string{"jian.bian@shopee.com", "hanwen.liu@shopee.com"}

		// send email
		if err = emailUtil.SendMail(srv.config.SMTPAddress, srv.config.MailSender, mailSubject, mailBody, recipients, ccs, srv.config.User, srv.config.Pwd); err != nil {
			log.Errorf("RankReportService sendNewFingerRankToBizUsers SendMail error:%s", err.Error())
			continue
		}
		if err = srv.saveNewFingerReportLog(taskUUID, newFinger, subRecord[0].ProductLine, subRecord[0].Owners, subRecord[0].Leaders, subRecord.NewFingerCount(), subRecord.NewSqlQueryCount(), dailyReport.ReportDayTime); err != nil {
			log.Errorf("RankReportService saveNewFingerReportLog saveReportLog error:%s", err.Error())
			continue
		}
	}
	return nil
}

func (srv *RankReportService) isSend(day, uuid string) (has bool) {
	var (
		err       error
		reportLog *mysql.DailyReportLog
	)

	if reportLog, err = mysql.FindDailyReportLog(srv.mysqlDB, uuid, day); err != nil {
		log.Errorf("RankReportService sendDBRankToInternal FindDailyReportLog error:%s", err.Error())
		return false
	}

	if reportLog != nil && len(reportLog.TaskUUID) > 0 {
		has = true
	}
	return has
}

func (srv *RankReportService) newFingerReportIsSend(day, uuid string) (has bool) {
	var (
		err       error
		reportLog *mysql.DailyNewFingerReportLog
	)

	if reportLog, err = mysql.FindDailyNewFingerReportLog(srv.mysqlDB, uuid, day); err != nil {
		log.Errorf("RankReportService sendDBRankToInternal FindDailyReportLog error:%s", err.Error())
		return false
	}

	if reportLog != nil && len(reportLog.TaskUUID) > 0 {
		has = true
	}
	return has
}

func (srv *RankReportService) saveReportLog(uuid, taskName, productLine string, owners, leaders []string, reportDay time.Time) error {
	reportLog := &mysql.DailyReportLog{
		TaskUUID:    uuid,
		TaskName:    taskName,
		ProductLine: productLine,
		DBEnv:       srv.config.Env,
		Owners:      strings.Join(owners, ","),
		Leaders:     strings.Join(leaders, ","),
		ReportDay:   reportDay,
	}

	return mysql.SaveDailyReportLog(srv.mysqlDB, reportLog)
}

func (srv *RankReportService) saveNewFingerReportLog(uuid, taskName, productLine string, owners, leaders []string, newFingerCount, newSqlQueryCount int, reportDay time.Time) error {
	reportLog := &mysql.DailyNewFingerReportLog{
		TaskUUID:    uuid,
		TaskName:    taskName,
		ProductLine: productLine,
		DBEnv:       srv.config.Env,
		Owners:      strings.Join(owners, ","),
		Leaders:     strings.Join(leaders, ","),
		NewFinger:   newFingerCount,
		NewSqlQuery: newSqlQueryCount,
		ReportDay:   reportDay,
	}

	return mysql.SaveDailyNewFingerReportLog(srv.mysqlDB, reportLog)
}

// db_rank_detail_url = "https://space.test.shopee.io/console/rdslivetest/rds/slow_query/live?cmdbService=%s&dbEnv=%s&clusterUUID=%s&dbNames=%s&startTime=%d&endTime=%d"
func (srv *RankReportService) buildDBDetailUrl(dailyRank *modelReport.DBSlowQueryRankDaily) {
	for i := range dailyRank.Rank {
		dailyRank.Rank[i].DetailLink = fmt.Sprintf(srv.config.DBRankDetailUrl, dailyRank.Rank[i].OwnCMDB, dailyRank.DBEnv, dailyRank.Rank[i].ClusterUUID, dailyRank.Rank[i].DBName, timeUtil.StartOfTheDayStamp(dailyRank.RankDayTime).Unix(), timeUtil.EndOfTheDayStamp(dailyRank.RankDayTime).Unix())
	}
}

// finger_rank_detail_url = "https://space.test.shopee.io/console/rdslivetest/rds/slow_query/live?cmdbService=%s&dbEnv=%s&clusterUUID=%s&dbNames=%s&finger_id=%s&startTime=%d&endTime=%d"
func (srv *RankReportService) buildFingerDetailUrl(dailyRank *modelReport.FingerSlowQueryRankDaily) {
	for i := range dailyRank.Rank {
		dailyRank.Rank[i].DetailLink = fmt.Sprintf(srv.config.FingerRankDetailUrl, dailyRank.Rank[i].OwnCMDB, dailyRank.DBEnv, dailyRank.Rank[i].ClusterUUID, dailyRank.Rank[i].DBName, dailyRank.Rank[i].FingerID, timeUtil.StartOfTheDayStamp(dailyRank.RankDayTime).Unix(), timeUtil.EndOfTheDayStamp(dailyRank.RankDayTime).Unix())
	}
}

// new_finger_rank_detail_url = "https://space.test.shopee.io/console/rdslivetest/slow_query/live?cmdbService=%s&dbEnv=%s&clusterUUID=%s&dbNames=%s&startTime=%d&endTime=%d&appear_type=new_appeared"
func (srv *RankReportService) buildNewFingerDetailUrl(dailyRank *modelReport.NewFingerDailyReport) {
	for i := range dailyRank.NewFingerInfos {
		dailyRank.NewFingerInfos[i].DetailLink = fmt.Sprintf(srv.config.NewFingerReportDetailUrl, dailyRank.NewFingerInfos[i].OwnCMDB, dailyRank.DBEnv, dailyRank.NewFingerInfos[i].ClusterUUID, dailyRank.NewFingerInfos[i].DBName, timeUtil.StartOfTheDayStamp(dailyRank.ReportDayTime).Unix(), timeUtil.EndOfTheDayStamp(dailyRank.ReportDayTime).Unix())
	}
}

func buildTaskUUID(reportName, productLine, day string, keys []string) string {
	encoder := md5.New()
	str := reportName + productLine + day + strings.Join(keys, ",")
	encoder.Write([]byte(str))
	return hex.EncodeToString(encoder.Sum(nil))
}

func getDBNames(list []*modelReport.DBQueryTime) (names []string) {
	for _, one := range list {
		names = append(names, one.DBName)
	}
	return names
}

func getFingerNames(list []*modelReport.FingerQueryTime) (names []string) {
	for _, one := range list {
		names = append(names, one.DBName)
	}
	return names
}

func getNewFingerNames(list []*modelReport.NewFingerInfo) (names []string) {
	for _, one := range list {
		names = append(names, one.DBName)
	}
	return names
}
