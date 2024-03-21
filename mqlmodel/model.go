package mqlmodel

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type AuthConf struct {
	Username string
	Password string
}

type MqlModel interface {
	InstanceName() string
	DatabaseName() string
}

type MqlAuthModel interface {
	MqlModel
	GetAuthConf() AuthConf
}

type getClientFunc func() *client

var (
	dbInstanceClientFuncMap = make(map[string]getClientFunc)
)

func getMysqlDB(instance string) *gorm.DB {
	clientFunc, ok := dbInstanceClientFuncMap[instance]
	if !ok {
		panic(fmt.Sprintf("db instance %s not found", instance))
	}

	return clientFunc().DB()
}

func GetMysqlDByModel(m MqlModel) *gorm.DB {
	db := getMysqlDB(m.InstanceName()).Model(m)
	return db
}

func registerInstance(instance string, conf *config) {
	dbInstanceClientFuncMap[instance] = func() getClientFunc {
		var cli *client
		var once sync.Once

		f := func() *client {
			once.Do(func() {
				cli = NewClient(conf)
			})
			return cli
		}

		return f
	}()
}

func RegisterMqlAuthModel(autoMigrate bool, ms ...MqlAuthModel) {
	for _, m := range ms {
		if _, exists := dbInstanceClientFuncMap[m.InstanceName()]; exists {
			continue
		}
		conf := &config{
			instanceName: m.InstanceName(),
			database:     m.DatabaseName(),
			AuthConf:     m.GetAuthConf(),
		}

		registerInstance(m.InstanceName(), conf)
	}
}

func RegisterMqlModel(auth AuthConf, autoMigrate bool, ms ...MqlModel) {
	for _, m := range ms {
		if _, exists := dbInstanceClientFuncMap[m.InstanceName()]; exists {
			continue
		}
		conf := &config{
			instanceName: m.InstanceName(),
			database:     m.DatabaseName(),
			AuthConf:     auth,
		}

		registerInstance(m.InstanceName(), conf)

		if autoMigrate {
			GetMysqlDByModel(m).AutoMigrate(m)
		}
	}
}
