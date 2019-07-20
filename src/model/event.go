package model

import (
	"strings"
	"time"
)

type Event struct {
	Id           int64     `json:"id"`
	Sid          int64     `json:"sid"`
	Sname        string    `json:"sname"`
	NodePath     string    `json:"node_path"`
	Endpoint     string    `json:"endpoint"`
	Priority     int       `json:"priority"`
	EventType    string    `json:"event_type"` // alert|recovery
	Category     int       `json:"category"`
	Status       uint16    `json:"status"`
	HashId       uint64    `json:"hashid"  xorm:"hashid"`
	Etime        int64     `json:"etime"`
	Value        string    `json:"value"`
	Info         string    `json:"info"`
	Created      time.Time `json:"created" xorm:"created"`
	Detail       string    `json:"detail"`
	Users        string    `json:"users"`
	Groups       string    `json:"groups"`
	Nid          int64     `json:"nid"`
	NeedUpgrade  int       `json:"need_upgrade"`
	AlertUpgrade string    `json:"alert_upgrade"`
}

type EventDetail struct {
	Metric     string              `json:"metric"`
	Tags       map[string]string   `json:"tags"`
	Points     []*EventDetailPoint `json:"points"`
	PredPoints []*EventDetailPoint `json:"pred_points,omitempty"` // 预测值, 预测值不为空时, 现场值对应的是实际值
}

type EventDetailPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type EventAlertUpgrade struct {
	Users    string `json:"users"`
	Groups   string `json:"groups"`
	Duration int    `json:"duration"`
	Level    int    `json:"level"`
}

func ParseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}

type EventSlice []*Event

func (e EventSlice) Len() int {
	return len(e)
}

func (e EventSlice) Less(i, j int) bool {
	return e[i].Etime < e[j].Etime
}

func (e EventSlice) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func SaveEvent(event *Event) error {
	_, err := DB["mon"].Insert(event)
	return err
}

func SaveEventStatus(id int64, status string) error {
	sql := "update event set status = status | ? where id = ?"
	_, err := DB["mon"].Exec(sql, GetStatus(status), id)
	return err
}

func UpdateEventPriority(id int64, priority int) error {
	sql := "update event set priority=? where id=?"
	_, err := DB["mon"].Exec(sql, priority, id)

	return err
}

func (e *Event) GetEventDetail() ([]EventDetail, error) {
	detail := []EventDetail{}

	err := json.Unmarshal([]byte(e.Detail), &detail)
	return detail, err
}

func EventTotal(stime, etime int64, nodePath, query, eventType string, priorities, sendTypes []string) (int64, error) {
	session := DB["mon"].Where("etime > ? and etime < ? and node_path = ?", stime, etime, nodePath)
	if len(priorities) > 0 && priorities[0] != "" {
		session = session.In("priority", priorities)
	}

	if len(sendTypes) > 0 && sendTypes[0] != "" {
		session = session.In("status", GetFlagsByStatus(sendTypes))
	}

	if eventType != "" {
		session = session.Where("event_type=?", eventType)
	}

	if query != "" {
		fields := strings.Fields(query)
		for i := 0; i < len(fields); i++ {
			if fields[i] == "" {
				continue
			}

			q := "%" + fields[i] + "%"
			session = session.Where("sname like ? or endpoint like ? or node_path like ?", q, q, q)
		}
	}

	total, err := session.Count(new(Event))
	return total, err
}

func EventGets(stime, etime int64, nodePath, query, eventType string, priorities, sendTypes []string, limit, offset int) ([]Event, error) {
	var objs []Event

	session := DB["mon"].Where("etime > ? and etime < ? and node_path = ?", stime, etime, nodePath)
	if len(priorities) > 0 && priorities[0] != "" {
		session = session.In("priority", priorities)
	}

	if len(sendTypes) > 0 && sendTypes[0] != "" {
		session = session.In("status", GetFlagsByStatus(sendTypes))
	}

	if eventType != "" {
		session = session.Where("event_type=?", eventType)
	}

	if query != "" {
		fields := strings.Fields(query)
		for i := 0; i < len(fields); i++ {
			if fields[i] == "" {
				continue
			}

			q := "%" + fields[i] + "%"
			session = session.Where("sname like ? or endpoint like ? or node_path like ?", q, q, q)
		}
	}

	err := session.Desc("etime").Limit(limit, offset).Find(&objs)

	return objs, err
}

func EventGet(col string, value interface{}) (*Event, error) {
	var obj Event
	has, err := DB["mon"].Where(col+"=?", value).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func DelEventOlder(ts int64, batch int) error {
	sql := "delete from event where etime < ? limit ?"
	_, err := DB["mon"].Exec(sql, ts, batch)

	return err
}

func EventAlertUpgradeUnMarshal(str string) (EventAlertUpgrade, error) {
	var obj EventAlertUpgrade
	if strings.TrimSpace(str) == "" {
		return EventAlertUpgrade{
			Users:    "[]",
			Groups:   "[]",
			Duration: 0,
			Level:    0,
		}, nil
	}

	err := json.Unmarshal([]byte(str), &obj)
	return obj, err
}
