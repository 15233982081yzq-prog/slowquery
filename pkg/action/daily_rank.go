package action

import (
	timeUtil "smart-slowquery/internal/util/time"
	rankService "smart-slowquery/pkg/service/statistics/daily"

	"smart-slowquery/internal/model/mysql"
	"smart-slowquery/internal/model/report"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/dbms"
	"smart-slowquery/pkg/store/response"

	"errors"
	"time"
)

var orderBy = "query_time"
var newFingerTop = 20

type ReportRankAction struct {
	rankSrv *rankService.SlowQueryRankService
	dbmsSrv *dbms.DBMetaService
}

func NewReportRankAction(rankService *rankService.SlowQueryRankService, dbmsService *dbms.DBMetaService) *ReportRankAction {
	return &ReportRankAction{
		rankSrv: rankService,
		dbmsSrv: dbmsService,
	}
}

func (act *ReportRankAction) DBDailyRank(day time.Time, dbEnv string, top int) (dailyRank *report.DBSlowQueryRankDaily, err error) {
	var rank []*mysql.DBSlowQueryDailyRank

	if rank, err = act.rankSrv.DBSlowQueryRank(orderBy, dbEnv, day, top); err != nil {
		log.Errorf("ReportRankAction  order_by:%s ,error:%s", orderBy, err.Error())
		return nil, err
	}
	if len(rank) == 0 {
		log.Warningf("ReportRankAction day:%s ,db rank top:%d is empty", day.GoString(), top)
		return nil, errors.New("db rank is empty")
	}

	log.Infof("ReportRankAction dbDailyRank size:%d", len(rank))
	dailyRank = &report.DBSlowQueryRankDaily{
		OrderBy:     orderBy,
		RankDay:     day.Format(timeUtil.DayFormat),
		RankDayTime: rank[0].RankDay,
		DBEnv:       rank[0].DBEnv,
	}

	for idx, _ := range rank {
		slowQueryRank := rank[idx]
		ownCMDB, productLine, owner, leaders, err := act.dbmsSrv.GetOwnShip(slowQueryRank.DBName, slowQueryRank.DBEnv)
		if err != nil {
			log.Warningf("ReportRankAction dbDailyRank cluster_uuid:%s ,database_name:%s, env:%s GetOwnShip error:%s", slowQueryRank.ClusterUUID, slowQueryRank.DBName, slowQueryRank.DBEnv, err.Error())
			continue
		}

		queryTime := &report.DBQueryTime{
			SerialNo:    1 + idx,
			ClusterUUID: slowQueryRank.ClusterUUID,
			DBName:      slowQueryRank.DBName,
			OwnCMDB:     ownCMDB,
			ProductLine: productLine,
			Owners:      []string{owner},
			Leaders:     leaders,
			Total:       slowQueryRank.RankScore,
			Count:       slowQueryRank.SqlCount,
			WeekOnWeek:  slowQueryRank.WeekOnWeek,
			DetailLink:  "",
		}
		dailyRank.Rank = append(dailyRank.Rank, queryTime)
	}
	log.Infof("ReportRankAction SendDBRankReport dbDailyRank size:%d", len(rank))

	return dailyRank, nil
}

func (act *ReportRankAction) FingerDailyRank(day time.Time, dbEnv string, top int) (fingerRank *report.FingerSlowQueryRankDaily, err error) {
	var rank []*mysql.FingerSlowQueryDailyRank

	if rank, err = act.rankSrv.FingerSlowQueryRank(orderBy, dbEnv, day, top); err != nil {
		log.Errorf("ReportRankAction fingerDailyRank order_by:%s ,error:%s", orderBy, err.Error())
		return nil, err
	}
	if len(rank) == 0 {
		log.Warningf("ReportRankAction day:%s ,finger rank top:%d is empty", day.GoString(), top)
		return nil, errors.New("finger rank is empty")
	}

	log.Infof("ReportRankAction fingerDailyRank size:%d", len(rank))

	fingerRank = &report.FingerSlowQueryRankDaily{
		OrderBy:     orderBy,
		RankDay:     day.Format(timeUtil.DayFormat),
		RankDayTime: rank[0].RankDay,
		DBEnv:       rank[0].DBEnv,
	}

	for idx, _ := range rank {
		slowQueryRank := rank[idx]
		ownCMDB, productLine, owner, leaders, err := act.dbmsSrv.GetOwnShip(slowQueryRank.DBName, slowQueryRank.DBEnv)
		if err != nil {
			log.Warningf("ReportRankAction fingerDailyRank cluster_uuid:%s ,database_name:%s, env:%s GetOwnShip error:%s", slowQueryRank.ClusterUUID, slowQueryRank.DBName, slowQueryRank.DBEnv, err.Error())
			continue
		}

		fingerQueryTime := &report.FingerQueryTime{
			SerialNo:    1 + idx,
			FingerID:    slowQueryRank.FingerID,
			FingerSql:   slowQueryRank.FingerSql,
			ClusterUUID: slowQueryRank.ClusterUUID,
			DBName:      slowQueryRank.DBName,
			OwnCMDB:     ownCMDB,
			ProductLine: productLine,
			Owners:      []string{owner},
			Leaders:     leaders,
			Total:       slowQueryRank.RankScore,
			Count:       slowQueryRank.SqlCount,
			WeekOnWeek:  slowQueryRank.WeekOnWeek,
			DetailLink:  "",
		}
		fingerRank.Rank = append(fingerRank.Rank, fingerQueryTime)
	}
	return fingerRank, nil
}

func (act *ReportRankAction) NewFingerDailyReportRecords(day time.Time, dbEnv string) (*report.NewFingerDailyReport, error) {
	var (
		record                      []*response.NewFingerReportRecord
		ownCMDB, productLine, owner string
		leaders                     []string
		err                         error
		reportRecords               = &report.NewFingerDailyReport{OrderBy: orderBy, ReportDay: day.Format(timeUtil.DayFormat), DBEnv: dbEnv, ReportDayTime: day, Top: newFingerTop}
	)

	reportRecords.NewFingerInfos = make([]*report.NewFingerInfo, 0)
	if record, err = act.rankSrv.NewFingerReportList(orderBy, dbEnv, reportRecords.Top, day); err != nil {
		log.Errorf("ReportRankAction NewFingerReport order_by:%s ,error:%s", orderBy, err.Error())
		return nil, err
	}

	for _, newFinger := range record {
		ownCMDB, productLine, owner, leaders, err = act.dbmsSrv.GetOwnShip(newFinger.DatabaseName, dbEnv)
		if err != nil {
			log.Warningf("ReportRankAction NewFingerDailyReportRecords database_name:%s, env:%s GetOwnShip error:%s", newFinger.DatabaseName, dbEnv, err.Error())
			continue
		}
		reportRecords.NewFingerInfos = append(reportRecords.NewFingerInfos, &report.NewFingerInfo{
			DBName:      newFinger.DatabaseName,
			ClusterUUID: newFinger.ClusterUUID,
			OwnCMDB:     ownCMDB,
			ProductLine: productLine,
			Owners:      []string{owner},
			Leaders:     leaders,
			NewFinger:   newFinger.QueryFingerCount,
			NewSqlQuery: newFinger.QuerySQLCount,
			Datetime:    newFinger.Datetime.Format(timeUtil.SecondFormat),
		})
	}

	log.Infof("ReportRankAction NewFingerDailyReportRecords size:%d", len(reportRecords.NewFingerInfos))
	return reportRecords, nil
}
