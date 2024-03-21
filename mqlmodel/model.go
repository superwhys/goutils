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

func registerAndMigrate(instance, database string, auth AuthConf, ms ...MqlModel) {
	if _, exists := getInstanceClientFunc(instance, database); !exists {
		conf := &config{
			instanceName: instance,
			database:     database,
			AuthConf:     auth,
		}
		registerInstance(instance, database, conf)
	}
}

func validateModels(ms ...MqlModel) (instance string, database string) {
	for idx, m := range ms {
		if idx == 0 {
			instance = m.InstanceName()
			database = m.DatabaseName()
		} else if instance != m.InstanceName() || database != m.DatabaseName() {
			panic(fmt.Sprintf("model instance: %s-%s not equal with other model: %s-%s", m.InstanceName(), m.DatabaseName(), instance, database))
		}
	}
	return instance, database
}

func RegisterMqlAuthModel(ms ...MqlAuthModel) {
	instance, database := validateModels(toInterfaceSlice(ms)...)
	var auth AuthConf
	if len(ms) > 0 {
		auth = ms[0].GetAuthConf()
	}
	registerAndMigrate(instance, database, auth, toInterfaceSlice(ms)...)
}

func RegisterMqlModel(auth AuthConf, autoMigrate bool, ms ...MqlModel) {
	instance, database := validateModels(ms...)
	registerAndMigrate(instance, database, auth, ms...)
}

func toInterfaceSlice[T MqlModel](ms []T) []MqlModel {
	result := make([]MqlModel, len(ms))
	for i, m := range ms {
		result[i] = m
	}

	return result
}
