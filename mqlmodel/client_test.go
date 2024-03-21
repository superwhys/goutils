package mqlmodel

import (
	"testing"

	"github.com/superwhys/goutils/lg"
)

type UserModel struct {
	ID   uint `gorm:"primarykey"`
	Name string
	Age  int
}

func (um *UserModel) TableName() string {
	return "user"
}

func TestDialDB(t *testing.T) {
	auth := AuthConf{
		Instance: "localhost:3306",
		Database: "sql_test",
		Username: "root",
		Password: "yang4869",
	}

	RegisterMqlModel(auth, &UserModel{})
	var resp []*UserModel
	if err := GetMysqlDByModel(&UserModel{}).Find(&resp).Error; err != nil {
		lg.Errorf("get user data error: %v", err)
		return
	}

	lg.Info(lg.Jsonify(resp))
}
