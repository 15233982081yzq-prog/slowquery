package alert

import (
	timeUtil "smart-slowquery/internal/util/time"
	storeMsql "smart-slowquery/pkg/store/mysql"

	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/log"

	"errors"
	"time"
)

const (
	NormalSoftStatus = "normal"
	DeleteSoftStatus = "delete"
)

type RuleTab struct {
	ID                 int       `gorm:"column:id"`
	AlertRuleUUID      string    `gorm:"column:alert_rule_uuid"`
	ChannelUUID        string    `gorm:"column:channel_uuid"`
	StrategyID         string    `gorm:"column:strategy_id"`
	CMDB               string    `gorm:"column:cmdb"`
	DBEnv              string    `gorm:"column:db_env"`
	Trigger            string    `gorm:"column:rule_trigger"`
	MonitorRuleID      string    `gorm:"column:monitor_rule_id"`
	AlertDisplayName   string    `gorm:"column:alert_display_name"`
	TemplateName       string    `gorm:"column:template_name"`
	Severity           string    `gorm:"column:severity"`
	PromQL             string    `gorm:"column:prom_ql"`
	Expression         string    `gorm:"column:expression"`
	ExpressionValue    int       `gorm:"column:expression_value"`
	ForRange           string    `gorm:"column:for_range"`
	EvaluationInterval string    `gorm:"column:evaluation_interval"`
	AlarmMsg           string    `gorm:"column:alarm_msg"`
	ResolveMsg         string    `gorm:"column:resolve_msg"`
	Status             string    `gorm:"column:rule_status"`
	SoftStatus         string    `gorm:"column:soft_status"`
	DbsJson            string    `gorm:"column:dbs_json"`
	Creator            string    `gorm:"column:creator"`
	Modifier           string    `gorm:"column:modifier"`
	ChannelType        string    `gorm:"column:channel_type"`
	CreateTime         time.Time `gorm:"column:create_time"`
	UpdateTime         time.Time `gorm:"column:update_time"`
}

func (r *RuleTab) TableName() string {
	return "alert_rule_tab"
}

func (r *RuleTab) Save(db storeMsql.DB) error {
	return db.Save(r)
}

func (r *RuleTab) UpdateByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&RuleTab{}).Where(RuleTab{AlertRuleUUID: r.AlertRuleUUID}).Updates(r).Error
}

func (r *RuleTab) UpdateStatusByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&RuleTab{}).Where(RuleTab{AlertRuleUUID: r.AlertRuleUUID}).Updates(&RuleTab{Status: r.Status, Modifier: r.Modifier}).Error
}

func (r *RuleTab) DeleteByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&RuleTab{}).Where(RuleTab{AlertRuleUUID: r.AlertRuleUUID}).Delete(&RuleTab{}).Error
}

func (r *RuleTab) UpdateRuleSoftStatusToDeleteByUUID(db storeMsql.DB) error {
	return db.GetDBConn().Model(&RuleTab{}).Where(RuleTab{AlertRuleUUID: r.AlertRuleUUID}).Updates(map[string]interface{}{
		"soft_status": r.SoftStatus,
	}).Error
}

func (r *RuleTab) FindAlertRuleByUuid(db storeMsql.DB, uuid string) (*RuleTab, error) {
	res := &RuleTab{}
	if err := db.GetDBConn().Model(&RuleTab{}).Where(RuleTab{AlertRuleUUID: uuid}).First(res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func FindAlertRuleByUUID(db storeMsql.DB, uuid string) (ar *RuleTab, err error) {
	res := &RuleTab{}
	if err = db.GetDBConn().Model(&RuleTab{}).Where("alert_rule_uuid=?", uuid).First(&res).Error; err != nil {
		log.Errorf("FindAlertRuleByUUID error:%s", err.Error())
		return nil, err
	}
	return res, nil
}

func FindAlertCountByCond(db storeMsql.DB, cond *request.GetAlertRuleListRequest) (count int64, err error) {
	dbQuery := db.GetDBConn().Model(&RuleTab{})
	dbQuery = dbQuery.Where("soft_status=?", NormalSoftStatus)
	if cond.Name != "" {
		dbQuery = dbQuery.Where("alert_display_name=?", cond.Name)
	}

	if cond.Severity != "" {
		dbQuery = dbQuery.Where("severity=?", cond.Severity)
	}

	if cond.DBEnv != "" {
		dbQuery = dbQuery.Where("db_env=?", cond.DBEnv)
	}

	if cond.AlertTemplateName != "" {
		dbQuery = dbQuery.Where("template_name=?", cond.AlertTemplateName)
	}

	if cond.Creator != "" {
		dbQuery = dbQuery.Where("creator=?", cond.Creator)
	}

	if len(cond.CMDBS) > 0 {
		dbQuery = dbQuery.Where("cmdb in ?", cond.CMDBS)
	}

	if cond.Status != "" {
		dbQuery = dbQuery.Where("rule_status = ?", cond.Status)
	}

	if cond.StartTime > 0 && cond.EndTime > 0 {
		createTimeStart := timeUtil.UnixTimeFormat(int64(cond.StartTime), timeUtil.SecondFormat)
		createTimeEnd := timeUtil.UnixTimeFormat(int64(cond.EndTime), timeUtil.SecondFormat)
		dbQuery = dbQuery.Where("create_time BETWEEN ? AND ?", createTimeStart, createTimeEnd)
	}
	if err = dbQuery.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func FindAlertWithChannelByCond(db storeMsql.DB, cond *request.GetAlertRuleListRequest) (ruleAndChannel []*RuleAndChannel, err error) {
	res := make([]*RuleAndChannel, 0)
	dbQuery := db.GetDBConn().Model(&RuleTab{}).Select("*").Joins("left join alert_channel_tab on alert_rule_tab.channel_uuid = alert_channel_tab.channel_uuid")
	dbQuery = dbQuery.Where("alert_rule_tab.soft_status=?", NormalSoftStatus)
	if cond.Name != "" {
		dbQuery = dbQuery.Where("alert_rule_tab.alert_display_name=?", cond.Name)
	}

	if cond.Severity != "" {
		dbQuery = dbQuery.Where("alert_rule_tab.severity=?", cond.Severity)
	}

	if cond.AlertTemplateName != "" {
		dbQuery = dbQuery.Where("alert_rule_tab.template_name=?", cond.AlertTemplateName)
	}

	if cond.Creator != "" {
		dbQuery = dbQuery.Where("alert_rule_tab.creator=?", cond.Creator)
	}

	if cond.DBEnv != "" {
		dbQuery = dbQuery.Where("alert_rule_tab.db_env=?", cond.DBEnv)
	}

	if len(cond.CMDBS) > 0 {
		dbQuery = dbQuery.Where("alert_rule_tab.cmdb in ?", cond.CMDBS)
	}

	if cond.Status != "" {
		dbQuery = dbQuery.Where("alert_rule_tab.rule_status = ?", cond.Status)
	}

	if cond.StartTime > 0 && cond.EndTime > 0 {
		createTimeStart := timeUtil.UnixTimeFormat(int64(cond.StartTime), timeUtil.SecondFormat)
		createTimeEnd := timeUtil.UnixTimeFormat(int64(cond.EndTime), timeUtil.SecondFormat)
		dbQuery = dbQuery.Where("alert_rule_tab.create_time BETWEEN ? AND ?", createTimeStart, createTimeEnd)
	}
	if err = dbQuery.Order("alert_rule_tab.create_time desc").Limit(cond.PageSize).Offset((cond.Page - 1) * cond.PageSize).Find(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func FindAlertRuleTypeCount(db storeMsql.DB, cmdbs []string) (int, int, error) {
	var resp []struct {
		EnableRules  int `gorm:"column:enable_rules"`
		DisableRules int `gorm:"column:disable_rules"`
	}

	dbQuery := db.GetDBConn().Model(RuleTab{}).Select("COUNT(IF(rule_status = 'enable', 1, NULL)) AS enable_rules,COUNT(IF(rule_status = 'disable', 1, NULL)) AS disable_rules").
		Where("soft_status = ?", NormalSoftStatus)

	if len(cmdbs) > 0 {
		dbQuery = dbQuery.Where("cmdb in ?", cmdbs)
	}
	if err := dbQuery.Find(&resp).Error; err != nil {
		return 0, 0, err
	}

	if len(resp) < 1 {
		return 0, 0, errors.New("can`t find count info")
	}

	return resp[0].EnableRules, resp[0].DisableRules, nil
}
