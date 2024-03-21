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

func (um *UserModel) InstanceName() string {
	return "localhost:3306"
}

func (um *UserModel) DatabaseName() string {
	return "sql_test"
}

func (um *UserModel) GetAuthConf() AuthConf {
	return AuthConf{
		Username: "root",
		Password: "yang4869",
	}
}

func TestDialDB(t *testing.T) {
	RegisterMqlAuthModel(&UserModel{})
	var resp []*UserModel
	if err := GetMysqlDByModel(&UserModel{}).Find(&resp).Error; err != nil {
		lg.Errorf("get user data error: %v", err)
		return
	}

	lg.Info(lg.Jsonify(resp))
}
