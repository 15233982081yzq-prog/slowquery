package response

import "smart-slowquery/thrid-party/soar-dev/advisor"

type SuggestInfoVo struct {
	HeuristicSuggest []*Rule `json:"heuristic_suggest"` // 启发式建议
	MysqlSuggest     []*Rule `json:"mysql_suggest"`     // MySQL 返回的 ERROR 信息
	IndexSuggest     []*Rule `json:"index_suggest"`     // 索引优化建议
}

type Option func(rule *Rule)

func WithOutCase() Option {
	return func(rule *Rule) {
		rule.Case = ""
	}
}

func ToRuleVo(rules map[string]advisor.Rule, options ...Option) []*Rule {
	ruleVo := make([]*Rule, 0, len(rules))
	for _, rule := range rules {
		tmp := &Rule{
			Item:    rule.Item,
			Summary: rule.Summary,
			Content: rule.Content,
			Case:    rule.Case,
		}
		for _, option := range options {
			option(tmp)
		}
		ruleVo = append(ruleVo, tmp)
	}
	return ruleVo
}
