package mysql

import (
	"database/sql"

	"gorm.io/gorm"
)

// DB indicates all generic database behaviours.
type DB interface {
	Close() error

	GetFirst(obj interface{}, where string, args ...interface{}) error
	GetLast(obj interface{}, where string, args ...interface{}) error
	// GetAll(objs interface{}) error
	Query(objs interface{}, where string, args ...interface{}) error
	QueryForUpdate(objs interface{}, where string, args ...interface{}) error
	QueryAndOrder(obj interface{}, orderCondition string, where string, args ...interface{}) error
	Exec(where string, args ...interface{}) (int64, error)

	Save(obj interface{}) error
	Create(obj interface{}) error
	Delete(obj interface{}) error
	FirstOrUpdate(obj interface{}, where ...interface{}) error
	FirstOrCreate(obj interface{}, query interface{}, args ...interface{}) error
	UpdateOneColumn(obj interface{}, colName string, colValue interface{}) error
	UpdateColumns(tableName string, whereCol string, whereColValue interface{}, updateCols interface{}) error
	UpdateByStruct(tableName string, whereStruct interface{}, updateStruct interface{}) error
	Count(obj interface{}, query interface{}, args ...interface{}) (int64, error)
	PagingQuery(obj interface{}, pageSize int, offset int, where string, args ...interface{}) error
	Joins(tableName string, selectCol string, joinStr string) (*sql.Rows, error)
	// DBWithGorm  related interface
	DBWithGorm
	DBTransaction
}

// DBWithGorm interface that free up gorm handlers.
type DBWithGorm interface {
	GetDBConn() *gorm.DB
	GetAll(conn *gorm.DB, objs interface{}) error
}

type DBTransaction interface {
	Transaction(fc func(m DB) error, opts ...*sql.TxOptions) error
	Begin(opts ...*sql.TxOptions) DB
	RollBack() DB
	Commit() DB
}
