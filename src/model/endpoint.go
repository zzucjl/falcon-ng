package model

import (
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/toolkits/pkg/str"
)

type Endpoint struct {
	Id    int64  `json:"id"`
	Ident string `json:"ident"`
	Alias string `json:"alias"`
}

func EndpointGet(col string, val interface{}) (*Endpoint, error) {
	var obj Endpoint
	has, err := DB["portal"].Where(col+"=?", val).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func (e *Endpoint) Update(cols ...string) error {
	_, err := DB["portal"].Where("id=?", e.Id).Cols(cols...).Update(e)
	return err
}

func EndpointTotal(query, batch, field string) (int64, error) {
	session := buildEndpointWhere(query, batch, field)
	return session.Count(new(Endpoint))
}

func EndpointGets(query, batch, field string, limit, offset int) ([]Endpoint, error) {
	session := buildEndpointWhere(query, batch, field).OrderBy(field).Limit(limit, offset)
	var objs []Endpoint
	err := session.Find(&objs)
	return objs, err
}

func buildEndpointWhere(query, batch, field string) *xorm.Session {
	session := DB["portal"].Table(new(Endpoint))

	if batch == "" && query != "" {
		q := "%" + query + "%"
		session = session.Where("ident like ? or alias like ?", q, q)
	}

	if batch != "" {
		endpoints := str.ParseCommaTrim(batch)
		if len(endpoints) > 0 {
			session = session.In(field, endpoints)
		}
	}

	return session
}

func EndpointImport(endpoints []string) error {
	count := len(endpoints)
	if count == 0 {
		return nil
	}

	session := DB["portal"].NewSession()
	defer session.Close()

	for i := 0; i < count; i++ {
		arr := strings.Split(endpoints[i], "::")

		ident := strings.TrimSpace(arr[0])
		alias := ""
		if len(arr) == 2 {
			alias = strings.TrimSpace(arr[1])
		}

		if ident == "" {
			continue
		}

		err := endpointImport(session, ident, alias)
		if err != nil {
			return err
		}
	}

	return nil
}

func endpointImport(session *xorm.Session, ident, alias string) error {
	var endpoint Endpoint
	has, err := session.Where("ident=?", ident).Get(&endpoint)
	if err != nil {
		return err
	}

	if has {
		endpoint.Alias = alias
		_, err = session.Where("ident=?", ident).Cols("alias").Update(endpoint)
	} else {
		_, err = session.Insert(Endpoint{Ident: ident, Alias: alias})
	}

	return err
}

func EndpointDel(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	bindings, err := NodeEndpointGetByEndpointIds(ids)
	if err != nil {
		return err
	}

	for i := 0; i < len(bindings); i++ {
		err = NodeEndpointUnbind(bindings[i].NodeId, bindings[i].EndpointId)
		if err != nil {
			return err
		}
	}

	if _, err := DB["portal"].In("id", ids).Delete(new(Endpoint)); err != nil {
		return err
	}

	return nil
}