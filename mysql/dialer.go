package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/service/finder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DefaultUserName = "root"
)

type DialOption struct {
	User     string
	Password string
	DBName   string
	Logger   logger.Interface
}

type OptionFunc func(*DialOption)

func WithAuth(user, pwd string) OptionFunc {
	return func(do *DialOption) {
		do.User = user
		do.Password = pwd
	}
}

func WithDBName(db string) OptionFunc {
	return func(do *DialOption) {
		do.DBName = db
	}
}

func WithLogger(log logger.Interface) OptionFunc {
	return func(do *DialOption) {
		do.Logger = log
	}
}

func packDialOption(opts ...OptionFunc) *DialOption {
	opt := &DialOption{}
	for _, o := range opts {
		o(opt)
	}

	if opt.User == "" {
		opt.User = DefaultUserName
	}

	return opt
}

func generateDSN(address string, opts ...OptionFunc) string {
	opt := packDialOption(opts...)
	dsn := fmt.Sprintf(
		"%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		opt.User,
		opt.Password,
		address,
		opt.DBName,
	)
	lg.Debugf("gorm dial dsn: %v", dsn)
	return dsn
}

func configDB(sqlDB *sql.DB) {
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func DialGorm(service string, opts ...OptionFunc) (*gorm.DB, error) {
	address := finder.GetServiceFinder().GetAddress(service)

	dsn := generateDSN(address, opts...)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	configDB(sqlDB)
	return db, nil
}

func DialMysql(service string, opts ...OptionFunc) (*sql.DB, error) {
	address := finder.GetServiceFinder().GetAddress(service)

	dsn := generateDSN(address, opts...)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "db ping")
	}

	configDB(db)
	return db, nil
}
