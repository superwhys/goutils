package localrepos

import (
	"database/sql"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/superwhys/goutils/lg"
)

var (
	r    *LocalRepos
	mock sqlmock.Sqlmock
)

type User struct {
	Name string
	Age  int
}

func (u *User) GetId() string {
	return u.Name
}

func TestMain(m *testing.M) {
	var (
		db  *sql.DB
		err error
	)
	db, mock, err = sqlmock.New()
	if err != nil {
		panic(err)
	}

	mysqlRepo, err := NewMysqlDataStoreWithDB(db, "mysql", "select name, age from test", &User{})
	if err != nil {
		panic(err)
	}
	rows := mock.NewRows([]string{"name", "age"}).AddRow("yong", 19).AddRow("hao", 20)
	mock.ExpectPrepare("select name, age from test").ExpectQuery().WillReturnRows(rows)

	go func() {
		time.Sleep(1 * time.Second)
		rows := mock.NewRows([]string{"name", "age"}).AddRow("yong", 19).AddRow("hao", 20).AddRow("wen", 21)
		mock.ExpectPrepare("select name, age from test").ExpectQuery().WillReturnRows(rows)
	}()

	r = NewLocalRepos(mysqlRepo, WithRefreshInterval(5*time.Second))
	r.Start()
	m.Run()
}

func TestLocalRepos_AllValues(t *testing.T) {
	tests := []struct {
		name       string
		sleep      time.Duration
		beforeWant []HashData
		want       []HashData
	}{
		{
			name: "test-get-data",
			want: []HashData{
				&User{Name: "yong", Age: 19},
				&User{Name: "hao", Age: 20},
			},
		},
		{
			name:  "test-get-data-after-refresh",
			sleep: 6 * time.Second,
			beforeWant: []HashData{
				&User{Name: "yong", Age: 19},
				&User{Name: "hao", Age: 20},
			},
			want: []HashData{
				&User{Name: "yong", Age: 19},
				&User{Name: "hao", Age: 20},
				&User{Name: "wen", Age: 21},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortHashData(tt.beforeWant)
			sortHashData(tt.want)

			if tt.beforeWant != nil {
				got := r.AllValues()
				sortHashData(got)

				if !reflect.DeepEqual(got, tt.beforeWant) {
					t.Errorf("LocalRepos.AllValues() = %v, want %v", lg.Jsonify(got), lg.Jsonify(tt.beforeWant))
				}
			}

			if tt.sleep != 0 {
				time.Sleep(tt.sleep)
			}

			got := r.AllValues()
			sortHashData(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LocalRepos.AllValues() = %v, want %v", lg.Jsonify(got), lg.Jsonify(tt.want))
			}
		})
	}
}

func sortHashData(data []HashData) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].GetId() < data[j].GetId()
	})
}
