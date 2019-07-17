package model

import (
	"log"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/modules/portal/config"
)

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Dispname string `json:"dispname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Im       string `json:"im"`
}

func InitRoot() {
	var u User
	has, err := DB["uic"].Where("username=?", "root").Get(&u)
	if err != nil {
		log.Fatalln("cannot query user[root]", err)
	}

	if has {
		return
	}

	u = User{
		Username: "root",
		Password: config.CryptoPass("falcon"),
		Dispname: "超管",
	}

	_, err = DB["uic"].Insert(&u)
	if err != nil {
		log.Fatalln("cannot insert user[root]")
	}

	logger.Info("user root init done")
}
