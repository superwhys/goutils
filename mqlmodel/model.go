package mqlmodel

import (
	"fmt"
	"reflect"
	"sync"

	"gorm.io/gorm"
)

type MqlModel interface {
	TableName() string
}

type AuthConf struct {
	Instance string
	Database string
	Username string
	Password string
}

func (auth AuthConf) GetInstanceKey() string {
	return fmt.Sprintf("%v-%v", auth.Instance, auth.Database)
}

type getClientFunc func() *client

var (
	dbInstanceClientFuncMap = make(map[string]getClientFunc)
	modelDbMap              = make(map[string]string)
)

func getMysqlDB(m MqlModel) *gorm.DB {
	rt := reflect.TypeOf(m)
	modelKey := fmt.Sprintf("%v-%v", rt.String(), m.TableName())

	key, exists := modelDbMap[modelKey]
	if !exists {
		panic(fmt.Sprintf("model %v has not been register", rt.String()))
	}

	clientFunc, ok := getInstanceClientFunc(key)
	if !ok {
		panic(fmt.Sprintf("db instance %v not found", key))
	}

	return clientFunc().DB()
}

func getInstanceClientFunc(key string) (getClientFunc, bool) {
	f, exists := dbInstanceClientFuncMap[key]
	return f, exists
}

func GetMysqlDByModel(m MqlModel) *gorm.DB {
	db := getMysqlDB(m).Model(m)
	return db
}

func registerInstance(conf *config) {
	key := conf.AuthConf.GetInstanceKey()
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

func RegisterMqlModel(auth AuthConf, ms ...MqlModel) {
	if _, exists := getInstanceClientFunc(auth.GetInstanceKey()); exists {
		panic(fmt.Sprintf("instance %v database %v has been register", auth.Instance, auth.Database))
	}

	conf := &config{
		AuthConf: auth,
	}
	registerInstance(conf)

	for _, m := range ms {
		rt := reflect.TypeOf(m)
		modelKey := fmt.Sprintf("%v-%v", rt.String(), m.TableName())
		if key, exists := modelDbMap[modelKey]; exists {
			panic(fmt.Sprintf("model %v has been register into %v", modelKey, key))
		}
		modelDbMap[modelKey] = auth.GetInstanceKey()
	}
}
