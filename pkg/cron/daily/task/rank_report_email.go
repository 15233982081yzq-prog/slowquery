package task

import (
	timeUtil "smart-slowquery/internal/util/time"

	"smart-slowquery/internal/model/report"
	"smart-slowquery/pkg/action"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/report/email"
)

var (
	orderBy = "query_time"
	top     = 10
	day     = timeUtil.YesterdayTime()
)

type SendReportEmailTask struct {
	env       string
	reportAct *action.ReportRankAction
	emailSrv  *email.RankReportService
}

func NewSendReportEmailTask(reportAct *action.ReportRankAction, emailService *email.RankReportService, env string) *SendReportEmailTask {
	return &SendReportEmailTask{
		reportAct: reportAct,
		emailSrv:  emailService,
		env:       env,
	}
}

func (tk *SendReportEmailTask) DO() (err error) {
	if err = tk.dbDailyRank(); err != nil {
		log.Errorf("SendReportEmailTask dbDailyRank error:%s", err.Error())
	}
	if err = tk.fingerDailyRank(); err != nil {
		log.Errorf("SendReportEmailTask  fingerDailyRank error:%s", err.Error())
	}
	if err = tk.newFingerDailyReport(); err != nil {
		log.Errorf("SendReportEmailTask newFingerDailyReport error:%s", err.Error())
	}
	return err
}

func (tk *SendReportEmailTask) Name() string {
	return "daily_rank_report_email"
}

func (tk *SendReportEmailTask) dbDailyRank() (err error) {
	var dailyRank *report.DBSlowQueryRankDaily

	if dailyRank, err = tk.reportAct.DBDailyRank(day, tk.env, top); err != nil {
		log.Errorf("ReportEmailTask  order_by:%s ,error:%v , rank:%v", orderBy, err, dailyRank)
		return err
	}

	log.Infof("ReportEmailTask SendDBRankReport dbDailyRank size:%d", len(dailyRank.Rank))

	return tk.emailSrv.SendDBRankReport(dailyRank)
}

func (tk *SendReportEmailTask) fingerDailyRank() (err error) {
	var fingerRank *report.FingerSlowQueryRankDaily

	if fingerRank, err = tk.reportAct.FingerDailyRank(day, tk.env, top); err != nil {
		log.Errorf("ReportEmailTask fingerDailyRank order_by:%s ,error:%v , rank:%v", orderBy, err, fingerRank)
		return
	}

	log.Infof("ReportEmailTask fingerDailyRank SendFingerRankReport size:%d", len(fingerRank.Rank))
	return tk.emailSrv.SendFingerRankReport(fingerRank)
}

func (tk *SendReportEmailTask) newFingerDailyReport() (err error) {
	var reportRecords *report.NewFingerDailyReport

	if reportRecords, err = tk.reportAct.NewFingerDailyReportRecords(day, tk.env); err != nil {
		log.Errorf("ReportEmailTask NewFingerDailyReportRecords,error:%v", err)
		return
	}

	log.Infof("ReportEmailTask NewFingerDailyReportRecords SendNewFingerRankReport size:%d", len(reportRecords.NewFingerInfos))
	return tk.emailSrv.SendNewFingerRankReport(reportRecords)
}
