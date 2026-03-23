package alert

import (
	storeMsql "smart-slowquery/pkg/store/mysql"
)

type RuleAndChannel struct {
	*RuleTab
	*ChannelTab
}

func SaveRulesAndChannelsAndStrategy(db storeMsql.DB, rules []*RuleTab, channels []*ChannelTab, strategies []*StrategyTab) error {
	return db.Transaction(func(db storeMsql.DB) error {
		for _, rule := range rules {
			if err := rule.Save(db); err != nil {
				return err
			}
		}
		for _, channel := range channels {
			if err := channel.Save(db); err != nil {
				return err
			}
		}
		for _, strategy := range strategies {
			if err := strategy.Save(db); err != nil {
				return err
			}
		}
		return nil
	})
}

func UpdateRuleAndChannelAndStrategy(db storeMsql.DB, rule *RuleTab, channel *ChannelTab, strategy *StrategyTab) error {
	return db.Transaction(func(db storeMsql.DB) error {
		if e := rule.UpdateByUUID(db); e != nil {
			return e
		}
		if e := channel.UpdateByUUID(db); e != nil {
			return e
		}
		if e := strategy.UpdateByStrategyID(db); e != nil {
			return e
		}
		return nil
	})
}

func UpdateRule(db storeMsql.DB, rule *RuleTab) error {
	return rule.UpdateByUUID(db)
}

func UpdateRuleStatus(db storeMsql.DB, rule *RuleTab) error {
	return rule.UpdateStatusByUUID(db)
}

func DeleteRuleAbout(db storeMsql.DB, rule *RuleTab, channel *ChannelTab, strategy *StrategyTab) error {
	return db.Transaction(func(db storeMsql.DB) error {
		if e := rule.DeleteByUUID(db); e != nil {
			return e
		}
		if e := channel.DeleteByUUID(db); e != nil {
			return e
		}
		if e := strategy.DeleteById(db); e != nil {
			return e
		}
		return nil
	})
}
