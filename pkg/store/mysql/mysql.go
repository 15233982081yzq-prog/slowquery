package mysql

import (
	"smart-slowquery/conf"
	"smart-slowquery/internal/util/errors"
	"smart-slowquery/pkg/log"

	"database/sql"
	"fmt"
	"strings"

	goErr "errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const MySQLDSNFmt string = "%s:%s@tcp(%s:%d)/%s?timeout=%ds&charset=utf8mb4&parseTime=True&loc=Local"

// MySQL indicates a MySQL connection.
type MySQL struct {
	conn *gorm.DB
}

// GetFirst gets the first row of matched data set.
func (m *MySQL) GetFirst(obj interface{}, where string, args ...interface{}) error {
	var err error
	if err = m.conn.Where(where, args...).First(obj).Error; goErr.Is(err, gorm.ErrRecordNotFound) {
		err = &errors.ErrRecordNotFound

	}
	return err
}

// GetLast gets the last row of matched data set.
func (m *MySQL) GetLast(obj interface{}, where string, args ...interface{}) error {
	var err error
	if err = m.conn.Where(where, args...).Last(obj).Error; goErr.Is(err, gorm.ErrRecordNotFound) {
		err = &errors.ErrRecordNotFound
	}
	return err
}

// GetAll gets all rows of the table.
//func (m *MySQL) GetAll(objs interface{}) error {
//	return m.conn.Find(objs).Error
//}

// Query queries data set.
func (m *MySQL) Query(obj interface{}, where string, args ...interface{}) error {
	return m.conn.Where(where, args...).Find(obj).Error
}

func (m *MySQL) QueryForUpdate(obj interface{}, where string, args ...interface{}) error {
	return m.conn.Clauses(clause.Locking{Strength: "UPDATE"}).Where(where, args...).Find(obj).Error
}

// QueryAndOrder queries data set and order.
func (m *MySQL) QueryAndOrder(obj interface{}, orderCondition string, where string, args ...interface{}) error {
	return m.conn.Where(where, args...).Order(orderCondition).Find(obj).Error
}

// Exec update / delete data object to database.
func (m *MySQL) Exec(where string, args ...interface{}) (int64, error) {
	db := m.conn.Exec(where, args...)
	return db.RowsAffected, db.Error
}

// Save updates a data object to database.
func (m *MySQL) Save(obj interface{}) error {
	return m.conn.Save(obj).Error
}

// Create new object to database.
func (m *MySQL) Create(obj interface{}) error {
	res := m.conn.Create(obj)
	return res.Error
}

// Delete object from database.
func (m *MySQL) Delete(obj interface{}) error {
	return m.conn.Delete(obj).Error
}

// Close closes the MySQL connection.
func (m *MySQL) Close() error {
	sqlDB, err := m.conn.DB() // m.conn.Close()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// FirstOrUpdate find first record, if not found, then update attr
func (m *MySQL) FirstOrUpdate(obj interface{}, where ...interface{}) error {
	return m.conn.FirstOrInit(obj, where).Error
}

// FirstOrCreate find first record, if not found, then create one
func (m *MySQL) FirstOrCreate(obj interface{}, query interface{}, args ...interface{}) error {
	return m.conn.Where(query, args).FirstOrCreate(obj).Error
}

// UpdateOneColumn update column
func (m *MySQL) UpdateOneColumn(obj interface{}, colName string, colValue interface{}) error {
	return m.conn.Model(obj).UpdateColumn(colName, colValue).Error
}

// UpdateColumns update columns
func (m *MySQL) UpdateColumns(tableName string, whereCol string, whereColValue interface{}, updates interface{}) error {
	return m.conn.Table(tableName).Where(fmt.Sprintf("%s in (?)", whereCol), whereColValue).Updates(updates).Error
}

// UpdateColumns update by struct

func (m *MySQL) UpdateByStruct(tableName string, whereStruct interface{}, updateStruct interface{}) error {
	res := m.conn.Table(tableName).Where(whereStruct).Updates(updateStruct)
	if res.Error == nil && res.RowsAffected <= 0 {
		return &errors.ErrZeroAffectedRow
	}
	return res.Error
}

// Count get count result
func (m *MySQL) Count(obj interface{}, query interface{}, args ...interface{}) (int64, error) {
	var count int64
	return count, m.conn.Model(obj).Where(query, args...).Count(&count).Error
}

// PagingQuery get Paging data
func (m *MySQL) PagingQuery(obj interface{}, pageSize int, offset int, where string, args ...interface{}) error {
	return m.conn.Limit(pageSize).Offset(offset).Where(where, args...).Find(obj).Error
}

// Joins table joins
func (m *MySQL) Joins(tableName string, selectCol string, joinStr string) (*sql.Rows, error) {
	return m.conn.Table(tableName).Select(selectCol).Joins(joinStr).Rows()
}

// =============================

// GetAll2 gets all rows of the table.
func (m *MySQL) GetAll(conn *gorm.DB, objs interface{}) error {
	return conn.Find(objs).Error
}

// GetDBConn expose the gorm.DB connection
func (m *MySQL) GetDBConn() *gorm.DB {
	return m.conn
}

func (m *MySQL) Transaction(fc func(m DB) error, opts ...*sql.TxOptions) error {
	return m.conn.Transaction(func(tx *gorm.DB) error {
		mysqlTx := &MySQL{conn: tx}
		return fc(mysqlTx)
	}, opts...)
}

func (m *MySQL) Begin(opts ...*sql.TxOptions) DB {
	mysqlTx := &MySQL{conn: m.conn.Begin(opts...)}
	return mysqlTx
}

func (m *MySQL) RollBack() DB {
	mysqlTx := &MySQL{conn: m.conn.Rollback()}
	return mysqlTx
}

func (m *MySQL) Commit() DB {
	mysqlTx := &MySQL{conn: m.conn.Commit()}
	return mysqlTx
}

// ConnectToMySQL connect to MySQL using config.MetaDBConnControl
func ConnectToMySQL(mdcc *conf.MetaDBConfig) (DB, error) {
	m := &MySQL{}
	var err error
	if strings.TrimSpace(mdcc.Username) == "" {
		return nil, &errors.ErrConfigDBUserNotSet
	}
	//if strings.TrimSpace(mdcc.Password) == "" {
	//	return nil, errors.ErrConfigDBPassNotSet
	//}
	if strings.TrimSpace(mdcc.Host) == "" {
		return nil, &errors.ErrConfigDBHostNotSet
	}
	if mdcc.Port <= 0 || mdcc.Port > 65535 {
		return nil, &errors.ErrConfigDBPortNotSet
	}
	if strings.TrimSpace(mdcc.DBName) == "" {
		return nil, &errors.ErrConfigDBNameNotSet
	}

	dsn := fmt.Sprintf(MySQLDSNFmt,
		mdcc.Username, mdcc.Password, mdcc.Host, mdcc.Port, mdcc.DBName, mdcc.ConnectTimeout)
	log.Infof("ConnectToMySQL dsn:%s", dsn)

	var logMode logger.Interface
	if mdcc.LogMode {
		logMode = logger.Default.LogMode(logger.Info)
	} else {
		logMode = logger.Default.LogMode(logger.Silent)
	}

	m.conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logMode,
	})
	if err != nil {
		log.Infof("gorm open dsn:%s , error:%s", dsn, err.Error())
		return nil, err
	}
	sqlDB, err := m.conn.DB()
	if err != nil {
		log.Infof("gorm conn.DB dsn:%s , error:%s", dsn, err.Error())
		return nil, err
	}
	sqlDB.SetMaxOpenConns(mdcc.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mdcc.MaxIdleConns)
	return m, err
}
