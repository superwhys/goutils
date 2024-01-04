package localrepos

import (
	"context"
	"reflect"

	"github.com/superwhys/goutils/qmgoutil"
	"gopkg.in/mgo.v2/bson"
)

type QmgoDataStore struct {
	ctx       context.Context
	client    *qmgoutil.Client
	hashModel HashData
	selector  bson.M
}

func NewQmgoDataStore(client *qmgoutil.Client, model HashData) *QmgoDataStore {
	_, isQmgoModel := model.(qmgoutil.QmgoModel)
	if !isQmgoModel {
		panic("model need to implement qmgoutil.QmgoModel interface as well")
	}

	return &QmgoDataStore{
		client:    client,
		hashModel: model,
		selector:  bson.M{},
	}
}

func (qds *QmgoDataStore) ReloadEntries(ctx context.Context, ch chan HashData) error {
	defer close(ch)

	cursor := qmgoutil.CollectionByClient(qds.client, qds.hashModel.(qmgoutil.QmgoModel)).Find(qds.ctx, qds.selector).Cursor()
	defer cursor.Close()

	var entryType reflect.Type
	if reflect.TypeOf(qds.hashModel).Kind() == reflect.Ptr {
		entryType = reflect.ValueOf(qds.hashModel).Elem().Type()
	} else {
		entryType = reflect.ValueOf(qds.hashModel).Type()
	}

	newQmgoModel := reflect.New(entryType).Interface().(HashData)
	for cursor.Next(newQmgoModel) {
		ch <- newQmgoModel

		newQmgoModel = reflect.New(entryType).Interface().(HashData)
	}

	return nil
}

func (qds *QmgoDataStore) Close() error {
	return qds.client.Client().Close(qds.ctx)
}
