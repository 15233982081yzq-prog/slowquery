package service

import (
	storeMsql "smart-slowquery/pkg/store/mysql"

	"smart-slowquery/conf"
	"smart-slowquery/internal/model/mysql"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/store/encryption"

	"fmt"
	"sync"
	"time"
)

type AccessService struct {
	Meta      *mysql.AccessMeta
	isUpdate  bool
	Password  string
	Key       string
	UserName  string
	rwLock    sync.RWMutex
	metaDBCfg *conf.MetaDBConfig
}

func NewAccessService(accessDB *conf.RemoteMysqlAccessConfig, metaDB *conf.MetaDBConfig) (*AccessService, error) {
	access := &AccessService{
		UserName:  accessDB.User,
		Password:  accessDB.TestPassword,
		Key:       accessDB.Key,
		isUpdate:  !accessDB.IsTest,
		metaDBCfg: metaDB,
	}
	return access, access.GetPasswdHash()
}

func (access *AccessService) CreateOrUpdateMySQLAccess(user string, password string) (err error) {
	log.Infof("accessService CreateOrUpdateMySQLAccess user:%s,password:%s", user, password)
	if !access.isUpdate {
		return nil
	}

	if user != access.UserName {
		return fmt.Errorf("CreateOrUpdateMySQLAccess param failed,user info inconsistent with config")
	}

	hash, err := encryption.AesEncryptCBC([]byte(password), []byte(access.Key))
	if err != nil {
		return err
	}

	s, err := storeMsql.NewSession(access.metaDBCfg.Username, access.metaDBCfg.Password, access.metaDBCfg.DBName, access.metaDBCfg.Host, access.metaDBCfg.Port)
	if err != nil {
		log.Errorf("access service init session error:%s", err.Error())
		return err
	}
	defer s.Close()

	currMeta := &mysql.AccessMeta{}
	if err := s.GetWithLimit(1, currMeta, &mysql.AccessMeta{UserName: user}); err != nil && err.Error() != "record not found" {
		return err
	}

	currMeta.UserName = user
	currMeta.PasswordHash = string(hash)
	currMeta.UpdateTime = time.Now()

	if err = s.Save(currMeta); err != nil {
		log.Errorf("CreateOrUpdateMySQLAccess user:%s ,error:%s", user, err.Error())
		return err
	}
	access.Password = password

	return nil
}

func (access *AccessService) GetPasswdHash() error {
	log.Infof("accessService UpdatePasswdHash")
	if !access.isUpdate {
		return nil
	}

	s, err := storeMsql.NewSession(access.metaDBCfg.Username, access.metaDBCfg.Password, access.metaDBCfg.DBName, access.metaDBCfg.Host, access.metaDBCfg.Port)
	if err != nil {
		log.Errorf("access service init session error:%s", err.Error())
		return err
	}
	defer s.Close()

	currMeta := &mysql.AccessMeta{}
	if err = s.GetWithLimit(1, currMeta, &mysql.AccessMeta{UserName: access.UserName}); err != nil {
		return fmt.Errorf("not found user in mysql access, user = %s ,error:%s", access.UserName, err.Error())
	}

	if currMeta.PasswordHash == "" {
		return fmt.Errorf(" user Password hash is empty, user = %s", access.UserName)
	}

	newPasswd, err := encryption.AesDecryptCBC([]byte(currMeta.PasswordHash), []byte(access.Key))
	if err != nil {
		return err
	}

	access.rwLock.Lock()
	defer access.rwLock.Unlock()

	access.Password = string(newPasswd)
	access.Meta = currMeta
	return nil
}

func (access *AccessService) GetPassword() string {
	access.rwLock.RLock()
	defer access.rwLock.RUnlock()
	return access.Password
}

func (access *AccessService) GetAccessMeta() (string, string, *mysql.AccessMeta) {
	access.rwLock.RLock()
	defer access.rwLock.RUnlock()
	return access.Password, access.Key, access.Meta
}
