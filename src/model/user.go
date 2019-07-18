package model

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/toolkits/pkg/logger"
	"gopkg.in/ldap.v3"

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

func LdapLogin(user, pass string) error {
	var conn *ldap.Conn
	var err error

	lc := config.Get().LDAP
	addr := fmt.Sprintf("%s:%d", lc.Host, lc.Port)

	if lc.TLS {
		conn, err = ldap.DialTLS("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = ldap.Dial("tcp", addr)
	}

	if err != nil {
		return fmt.Errorf("cannot dial ldap: %v", err)
	}

	defer conn.Close()

	if !lc.TLS && lc.StartTLS {
		err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return fmt.Errorf("ldap.conn startTLS fail: %v", err)
		}
	}

	err = conn.Bind(lc.BindUser, lc.BindPass)
	if err != nil {
		return fmt.Errorf("bind ldap fail: %v, use %s", err, lc.BindUser)
	}

	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.AuthFilter, user), // The filter to apply
		lc.Attributes,                    // A list attributes to retrieve
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("ldap search fail: %v", err)
	}

	if len(sr.Entries) == 0 {
		return fmt.Errorf("cannot find such user: %v", user)
	}

	if len(sr.Entries) > 1 {
		return fmt.Errorf("multi users is search, query user: %v", user)
	}

	err = conn.Bind(sr.Entries[0].DN, pass)
	if err != nil {
		return fmt.Errorf("password error")
	}

	cnt, err := DB["uic"].Where("username=?", user).Count(new(User))
	if err != nil {
		return err
	}

	if cnt > 0 {
		return nil
	}

	u := &User{
		Username: user,
		Password: "******",
		Dispname: "",
		Email:    "",
	}

	_, err = DB["uic"].Insert(u)
	return err
}

func PassLogin(user, pass string) error {
	var u User
	has, err := DB["uic"].Where("username=?", user).Cols("password").Get(&u)
	if err != nil {
		return err
	}

	if !has {
		return fmt.Errorf("user[%s] not found", user)
	}

	if config.CryptoPass(pass) != u.Password {
		return fmt.Errorf("password error")
	}

	return nil
}
