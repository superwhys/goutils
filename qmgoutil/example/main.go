package main

import (
	"context"
	"fmt"

	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/qmgoutil"
)

type User struct {
	Username string `bson:"username"`
}

func (u *User) DBConfig() qmgoutil.Config {
	return qmgoutil.Config{
		Address: "localhost:27017",
	}
}

func (u *User) DatabaseName() string {
	return "People"
}

func (u *User) CollectionName() string {
	return "users"
}

func main() {
	col := qmgoutil.CollectionByModel(&User{})
	col.InsertOne(context.Background(), &User{Username: "yanghw"})

	var res []*User
	if err := col.Find(context.Background(), &User{Username: "yanghw"}).All(&res); err != nil {
		panic(err)
	}
	fmt.Println(lg.Jsonify(res))
}
