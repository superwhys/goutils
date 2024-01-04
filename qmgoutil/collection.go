package qmgoutil

import "github.com/qiniu/qmgo"

func CollectionByModel(m QmgoModel) *qmgo.Collection {
	client := GetDBInstance(m)
	return client.Client().Database(m.DatabaseName()).Collection(m.CollectionName())
}

func CollectionByClient(client *Client, m QmgoModel) *qmgo.Collection {
	return client.Client().Database(m.DatabaseName()).Collection(m.CollectionName())
}
