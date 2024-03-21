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

func getMysqlDB(instance, database string) *gorm.DB {
	clientFunc, ok := getInstanceClientFunc(instance, database)
	if !ok {
		panic(fmt.Sprintf("db instance %s-%s not found", instance, database))
	}

	return clientFunc().DB()
}

func GetMysqlDByModel(m MqlModel) *gorm.DB {
	db := getMysqlDB(m.InstanceName(), m.DatabaseName()).Model(m)
	return db
}

func getInstanceClientFunc(instance, database string) (getClientFunc, bool) {
	key := fmt.Sprintf("%v-%v", instance, database)
	f, exists := dbInstanceClientFuncMap[key]
	return f, exists
}

func registerInstance(instance, database string, conf *config) {
	key := fmt.Sprintf("%v-%v", instance, database)
	dbInstanceClientFuncMap[key] = func() getClientFunc {
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
		if _, exists := getInstanceClientFunc(m.InstanceName(), m.DatabaseName()); exists {
			continue
		}
		conf := &config{
			instanceName: m.InstanceName(),
			database:     m.DatabaseName(),
			AuthConf:     m.GetAuthConf(),
		}

		registerInstance(m.InstanceName(), m.DatabaseName(), conf)
	}
}

func RegisterMqlModel(auth AuthConf, autoMigrate bool, ms ...MqlModel) {
	for _, m := range ms {
		if _, exists := getInstanceClientFunc(m.InstanceName(), m.DatabaseName()); exists {
			continue
		}
		conf := &config{
			instanceName: m.InstanceName(),
			database:     m.DatabaseName(),
			AuthConf:     auth,
		}

		registerInstance(m.InstanceName(), m.DatabaseName(), conf)

		if autoMigrate {
			GetMysqlDByModel(m).AutoMigrate(m)
		}
	}
}
