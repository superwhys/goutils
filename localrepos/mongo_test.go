package localrepos

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/qmgoutil"
)

var (
	qmgoStore *QmgoDataStore
)

type UserModel struct {
	Username string
	Age      int
}

func (um *UserModel) GetId() string {
	return um.Username
}

func (um *UserModel) DBConfig() qmgoutil.Config {
	return qmgoutil.Config{
		Address: "localhost:27017",
	}
}

func (um *UserModel) DatabaseName() string {
	return "person"
}

func (um *UserModel) CollectionName() string {
	return "users"
}

func init() {
	client := qmgoutil.NewClientWithModel(&UserModel{})
	qmgoStore = NewQmgoDataStore(client, &UserModel{})
	qmgoutil.CollectionByModel(&UserModel{}).InsertMany(context.Background(), []*UserModel{
		{Username: "yang", Age: 18},
		{Username: "hao", Age: 19},
		{Username: "wen", Age: 20},
	})
}

func TestQmgoDataStore_ReloadEntries(t *testing.T) {

	type args struct {
		ctx context.Context
		ch  chan HashData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{ctx: context.Background(), ch: make(chan HashData)}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qds := qmgoStore

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for data := range tt.args.ch {
					fmt.Println(lg.Jsonify(data))
				}
			}()

			if err := qds.ReloadEntries(tt.args.ctx, tt.args.ch); (err != nil) != tt.wantErr {
				t.Errorf("QmgoDataStore.ReloadEntries() error = %v, wantErr %v", err, tt.wantErr)
			}
			wg.Wait()
		})
	}
}
