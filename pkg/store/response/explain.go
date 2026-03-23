package response

import "database/sql"

const NullString = "NULL"

type ExplainResult struct {
	explains   []*ExplainInfo
	explainErr error
}

func NewExplainResult(explains []*ExplainInfo, err error) *ExplainResult {
	return &ExplainResult{
		explains:   explains,
		explainErr: err,
	}
}

func (er *ExplainResult) GetExplainInfos() []*ExplainInfo {
	return er.explains
}

func (er *ExplainResult) GetExplainError() error {
	return er.explainErr
}

type ExplainInfo struct {
	// gorm.Model
	ID            sql.NullString  `gorm:"Column:id"`
	SelectType    sql.NullString  `gorm:"Column:select_type"`
	Table         sql.NullString  `gorm:"Column:table"`
	Partitions    sql.NullString  `gorm:"Column:partitions"`
	Type          sql.NullString  `gorm:"Column:type"`
	PossibleKeys  sql.NullString  `gorm:"Column:possible_keys"`
	Key           sql.NullString  `gorm:"Column:key"`
	KeyLen        sql.NullString  `gorm:"Column:key_len"`
	Ref           sql.NullString  `gorm:"Column:ref"`
	Rows          sql.NullInt64   `gorm:"Column:rows"`
	Filtered      sql.NullFloat64 `gorm:"Column:filtered"`
	Extra         sql.NullString  `gorm:"Column:Extra"`
	Query         string
	StatementType string
}

func (explain *ExplainInfo) GetID() string {
	if explain.ID.Valid {
		return explain.ID.String
	}
	return NullString
}

func (explain *ExplainInfo) GetSelectType() string {
	if explain.SelectType.Valid {
		return explain.SelectType.String
	}
	return NullString
}

func (explain *ExplainInfo) GetTable() string {
	if explain.Table.Valid {
		return explain.Table.String
	}
	return NullString
}

func (explain *ExplainInfo) GetPartitions() string {
	if explain.Partitions.Valid {
		return explain.Partitions.String
	}
	return NullString
}

func (explain *ExplainInfo) GetType() string {
	if explain.Type.Valid {
		return explain.Type.String
	}
	return NullString
}

func (explain *ExplainInfo) GetPossibleKeys() string {
	if explain.PossibleKeys.Valid {
		return explain.PossibleKeys.String
	}
	return NullString
}

func (explain *ExplainInfo) GetKey() string {
	if explain.Key.Valid {
		return explain.Key.String
	}
	return NullString
}

func (explain *ExplainInfo) GetKeyLen() string {
	if explain.KeyLen.Valid {
		return explain.KeyLen.String
	}
	return NullString
}

func (explain *ExplainInfo) GetRef() string {
	if explain.Ref.Valid {
		return explain.Ref.String
	}
	return NullString
}

func (explain *ExplainInfo) GetExtra() string {
	if explain.Extra.Valid {
		return explain.Extra.String
	}
	return NullString
}
