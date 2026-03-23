package daily

import (
	sqlModel "smart-slowquery/internal/model/mysql"
	timeUtil "smart-slowquery/internal/util/time"
	storeMsql "smart-slowquery/pkg/store/mysql"

	"smart-slowquery/internal/model/mysql"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/request"
	"smart-slowquery/pkg/store/response"

	"fmt"
	"time"
)

type SlowQueryRankService struct {
	ck      store.CKStore
	mysqlDB storeMsql.DB
}

func NewSLowQueryRankService(store store.CKStore, mysqlDB storeMsql.DB) (*SlowQueryRankService, error) {
	if store == nil {
		return nil, fmt.Errorf("clickhouse store is nil")
	}
	if mysqlDB == nil {
		return nil, fmt.Errorf("mysql client is nil")
	}

	return &SlowQueryRankService{
		ck:      store,
		mysqlDB: mysqlDB,
	}, nil
}

func (srv *SlowQueryRankService) DBSlowQueryRank(order, dbEnv string, day time.Time, top int) ([]*mysql.DBSlowQueryDailyRank, error) {
	var (
		dailyRank, lastWeekDailyRank []*mysql.DBSlowQueryDailyRank
		lastWeekRankRecord           = make(map[string]int, 0)
		ckSlowRank                   *response.CKDBQueryRank
		dayParser                    time.Time
		err                          error
	)

	//优先从mysql库中获取慢日志日报
	if rank, err := srv.getDBDailyRankFromMysql(order, dbEnv, day, top); err == nil && len(rank) > 0 {
		log.Infof("srv.DBSlowQueryRank getDailyRankFromMysql rank size:%d", len(rank))
		return rank, nil
	}

	//从CK中统计慢日志日报
	if ckSlowRank, err = srv.getDailyDBRankFromCK(order, dbEnv, day, top); err != nil {
		log.Errorf("srv.DBSlowQueryRank getDailyRankFromCK order:%s ,day:%v ,top:%d ,error:%s", order, day, top, err.Error())
		return nil, err
	}
	if len(ckSlowRank.Rank) == 0 {
		log.Warningf("SlowQueryRankService DBSlowQueryRank db rank is empty")
		return nil, nil
	}
	log.Infof("ck ckSlowRank size:%d", len(ckSlowRank.Rank))

	//获取一周前的统计数据（from mysql）
	lastWeekDay := day.AddDate(0, 0, -7)
	if lastWeekDailyRank, err = srv.getDBDailyRankFromMysql(order, dbEnv, lastWeekDay, top); err == nil && len(lastWeekDailyRank) > 0 {
		log.Infof("srv.DBSlowQueryRank getLastWeekDailyRankFromMysql day:%v ,result size:%d", lastWeekDay, len(lastWeekDailyRank))
		for i := range lastWeekDailyRank {
			one := lastWeekDailyRank[i]
			lastWeekRankRecord[one.DBName] = one.SerialNo
		}
	} else {
		log.Warningf("srv.DBSlowQueryRank getLastWeekDailyRank from mysql,day:%v is nil", lastWeekDay)
	}

	for i := range ckSlowRank.Rank {
		ckOne := ckSlowRank.Rank[i]
		weekOneWeek := "New"
		if lastWeekSerialNo, ok := lastWeekRankRecord[ckOne.DBName]; ok {
			diff := i + 1 - lastWeekSerialNo
			switch {
			case diff == 0:
				weekOneWeek = "-"
			case diff > 0:
				weekOneWeek = fmt.Sprintf("Dropped:%d", diff)
			case diff < 0:
				weekOneWeek = fmt.Sprintf("Promoted:%d", 0-diff)
			}
		}

		rankOne := &mysql.DBSlowQueryDailyRank{
			SerialNo:    i + 1,
			ClusterUUID: ckOne.ClusterUUID,
			DBName:      ckOne.DBName,
			DBEnv:       ckSlowRank.Env,
			RankOrder:   ckSlowRank.OrderBy,
			RankScore:   ckOne.Total,
			SqlCount:    ckOne.Count,
			WeekOnWeek:  weekOneWeek,
			RankDay:     day,
		}
		dailyRank = append(dailyRank, rankOne)
	}

	dayParser, _ = time.Parse(timeUtil.DayFormat, timeUtil.UnixTimeFormat(time.Now().Unix(), timeUtil.DayFormat))
	//今日数据不入库
	if day.GoString() != dayParser.GoString() {
		if err := srv.saveDBDailyRank(dailyRank); err != nil {
			log.Warningf("srv.DBSlowQueryRank saveDailyRank to mysql error:%s", err.Error())
		}
	}

	return dailyRank, err
}

func (srv *SlowQueryRankService) FingerSlowQueryRank(order, dbEnv string, day time.Time, top int) ([]*mysql.FingerSlowQueryDailyRank, error) {
	var (
		dailyRank, lastWeekDailyRank []*mysql.FingerSlowQueryDailyRank
		lastWeekRankRecord           = make(map[string]int, 0)
		ckSlowRank                   *response.CKFingerQueryRank
		dayParser                    time.Time
		err                          error
	)

	//优先从mysql库中获取慢日志日报
	if rank, err := srv.getFingerDailyRankFromMysql(order, dbEnv, day, top); err == nil && len(rank) > 0 {
		log.Infof("srv.FingerSlowQueryRank getFingerDailyRankFromMysql rank size:%d", len(rank))
		return rank, err
	}

	//从CK中统计慢日志日报
	if ckSlowRank, err = srv.getDailyFingerRankFromCK(order, dbEnv, day, top); err != nil {
		log.Errorf("srv.FingerSlowQueryRank getDailyFingerRankFromCK order:%s ,day:%v ,top:%d ,error:%s", order, day, top, err.Error())
		return nil, err
	}
	if len(ckSlowRank.Rank) == 0 {
		log.Warningf("SlowQueryRankService FingerSlowQueryRank finger rank is empty")
		return nil, nil
	}
	log.Infof("srv.getDailyFingerRankFromCK rank size:%d", len(ckSlowRank.Rank))

	//获取一周前的统计数据（from mysql）
	lastWeekDay := day.AddDate(0, 0, -7)
	if lastWeekDailyRank, err = srv.getFingerDailyRankFromMysql(order, dbEnv, lastWeekDay, top); err == nil && len(lastWeekDailyRank) > 0 {
		log.Infof("srv.FingerSlowQueryRank getFingerDailyRankFromMysql day:%v ,result size:%d", lastWeekDay, len(lastWeekDailyRank))
		for i := range lastWeekDailyRank {
			one := lastWeekDailyRank[i]
			lastWeekRankRecord[one.DBName] = one.SerialNo
		}
	} else {
		log.Warningf("srv.FingerSlowQueryRank getLastWeekDailyRank from mysql,day:%v is nil", lastWeekDay)
	}

	for i := range ckSlowRank.Rank {
		ckOne := ckSlowRank.Rank[i]
		weekOneWeek := "New"
		if lastWeekSerialNo, ok := lastWeekRankRecord[ckOne.DBName]; ok {
			diff := i + 1 - lastWeekSerialNo
			switch {
			case diff == 0:
				weekOneWeek = "-"
			case diff > 0:
				weekOneWeek = fmt.Sprintf("Dropped:%d", diff)
			case diff < 0:
				weekOneWeek = fmt.Sprintf("Promoted:%d", 0-diff)
			}
		}

		rankOne := &mysql.FingerSlowQueryDailyRank{
			SerialNo:    i + 1,
			FingerID:    ckOne.FingerID,
			FingerSql:   ckOne.FingerSql,
			ClusterUUID: ckOne.ClusterUUID,
			DBName:      ckOne.DBName,
			DBEnv:       ckSlowRank.Env,
			RankOrder:   ckSlowRank.OrderBy,
			RankScore:   ckOne.Total,
			SqlCount:    ckOne.Count,
			WeekOnWeek:  weekOneWeek,
			RankDay:     day,
		}
		dailyRank = append(dailyRank, rankOne)
	}

	dayParser, _ = time.Parse(timeUtil.DayFormat, timeUtil.UnixTimeFormat(time.Now().Unix(), timeUtil.DayFormat))
	//今日数据不入库
	if day.GoString() != dayParser.GoString() {
		if err := srv.saveFingerDailyRank(dailyRank); err != nil {
			log.Warningf("srv.FingerSlowQueryRank saveFingerDailyRank to mysql error:%s", err.Error())
		}
	}

	return dailyRank, err
}

func (srv *SlowQueryRankService) NewFingerReportList(order, dbEnv string, top int, day time.Time) ([]*response.NewFingerReportRecord, error) {
	var (
		ckNewSlowReportRecords []*response.NewFingerReportRecord
		err                    error
	)
	//从CK中统计慢日志日报
	if ckNewSlowReportRecords, err = srv.getDailyNewFingerRecordFromCK(order, dbEnv, top, day); err != nil {
		log.Errorf("srv.NewFingerReport getDailyNewFingerRankFromCK order:%s ,day:%v ,error:%s", order, day, err.Error())
		return nil, err
	}
	log.Infof("srv.getDailyNewFingerRecordFromCK record size:%d", len(ckNewSlowReportRecords))
	return ckNewSlowReportRecords, err
}

func (srv *SlowQueryRankService) getDailyDBRankFromCK(order, dbEnv string, day time.Time, top int) (*response.CKDBQueryRank, error) {
	return srv.ck.GetDBSlowQueryRank(request.BuildSlowQueryRank(order, dbEnv, top, timeUtil.StartOfTheDayStamp(day), timeUtil.EndOfTheDayStamp(day)))
}

func (srv *SlowQueryRankService) getDailyFingerRankFromCK(order, dbEnv string, day time.Time, top int) (*response.CKFingerQueryRank, error) {
	return srv.ck.GetFingerSlowQueryRank(request.BuildSlowQueryRank(order, dbEnv, top, timeUtil.StartOfTheDayStamp(day), timeUtil.EndOfTheDayStamp(day)))
}

func (srv *SlowQueryRankService) getDailyNewFingerRecordFromCK(order, dbEnv string, top int, day time.Time) ([]*response.NewFingerReportRecord, error) {
	return srv.ck.GetNewFingerSlowQueryReportRecord(&request.SlowQueryNewFinger{
		OrderBy:       order,
		DBEnv:         dbEnv,
		StartTime:     timeUtil.StartOfTheDayStamp(day),
		Top:           top,
		EndTime:       timeUtil.EndOfTheDayStamp(day),
		IsNewAppeared: true,
	})
}

func (srv *SlowQueryRankService) saveDBDailyRank(rank []*mysql.DBSlowQueryDailyRank) (err error) {
	return sqlModel.SaveDBDailyRank(srv.mysqlDB, rank)
}

func (srv *SlowQueryRankService) getDBDailyRankFromMysql(order, dbEnv string, time time.Time, top int) ([]*mysql.DBSlowQueryDailyRank, error) {
	return sqlModel.FindDBDailyRank(srv.mysqlDB, order, time.Format(timeUtil.DayFormat), dbEnv, top)
}

func (srv *SlowQueryRankService) saveFingerDailyRank(rank []*mysql.FingerSlowQueryDailyRank) (err error) {
	return sqlModel.SaveFingerDailyRank(srv.mysqlDB, rank)
}

func (srv *SlowQueryRankService) getFingerDailyRankFromMysql(order, dbEnv string, time time.Time, top int) ([]*mysql.FingerSlowQueryDailyRank, error) {
	return sqlModel.FindFingerDailyRank(srv.mysqlDB, order, time.Format(timeUtil.DayFormat), dbEnv, top)
}
