package cron

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/json-iterator/go"
	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/cache"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/config"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/redi"
)

func ReadHighEvent() {
	queues := config.GetCfgYml().Queue.High
	if len(queues) == 0 {
		return
	}

	for {
		event, err := popEvent(queues)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		consume(event, true)
	}
}

func ReadLowEvent() {
	queues := config.GetCfgYml().Queue.Low
	if len(queues) == 0 {
		return
	}

	for {
		event, err := popEvent(queues)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		consume(event, false)
	}
}

func popEvent(queues []interface{}) (*model.Event, error) {
	queues = append(queues, 0)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	rc := redi.RedisConnPool.Get()
	defer rc.Close()

	reply, err := redis.Strings(rc.Do("BRPOP", queues...))
	if err != nil {
		if err != redis.ErrNil {
			logger.Warningf("get alarm event from redis failed, queues: %v, err: %v", queues, err)
		}
		return nil, err
	}

	if reply == nil {
		logger.Errorf("get alarm event from redis timeout")
		return nil, err
	}

	event := new(model.Event)
	if err = json.Unmarshal([]byte(reply[1]), event); err != nil {
		logger.Errorf("unmarshal redis reply failed, err: %v", err)
		return nil, err
	}

	stra, has := cache.StraCache.GetById(event.Sid)
	if !has {
		logger.Errorf("stra not found, stra id: %d, event: %+v", event.Sid, event)
		return nil, err
	}

	// 如果nid和endpoint的对应关系不正确，直接丢弃该event
	endpoint, err := model.EndpointGet("ident", event.Endpoint)
	if err != nil {
		logger.Errorf("get host_id failed, event: %+v, err: %v", event, err)
		return nil, err
	}

	nodePath := ""

	node, err := model.NodeGet("id", stra.Nid)
	if err != nil {
		logger.Errorf("get node failed, node id: %v, event: %+v, err: %v", stra.Nid, event, err)
		return nil, err
	}

	if node == nil {
		logger.Errorf("get node by id return nil, node id: %v, event: %+v", stra.Nid, event)
		return nil, fmt.Errorf("get node by id return nil")
	}

	nodePath = node.Path

	leafIds, err := node.LeafIds()
	if err != nil {
		logger.Errorf("get node leaf ids failed, node id: %v, event: %+v, err: %v", stra.Nid, event, err)
		return nil, err
	}

	nodeIds, err := model.NodeIdsGetByEndpointId(endpoint.Id)
	if err != nil {
		logger.Errorf("get node_endpoint by endpoint_id failed, event: %+v, err: %v", event, err)
		return nil, err
	}

	if nodeIds == nil || len(nodeIds) == 0 {
		logger.Errorf("endpoint(%s) not found, event: %+v", event.Endpoint, event)
		return nil, nil
	}

	has = false
	for i := 0; i < len(nodeIds); i++ {
		for j := 0; j < len(leafIds); j++ {
			if nodeIds[i] == leafIds[j] {
				has = true
				break
			}
		}
	}

	if !has {
		logger.Errorf("endpoint(%s) not match nid(%v), event: %+v", event.Endpoint, stra.Nid, event)
		return nil, nil
	}

	users, err := json.Marshal(stra.NotifyUser)
	if err != nil {
		logger.Errorf("users marshal failed, err: %v, event: %+v", err, event)
		return nil, err
	}

	groups, err := json.Marshal(stra.NotifyGroup)
	if err != nil {
		logger.Errorf("users marshal failed, err: %v, event: %+v", err, event)
		return nil, err
	}

	alertUpgrade, err := model.EventAlertUpgradeMarshal(stra.AlertUpgrade)
	if err != nil {
		logger.Errorf("EventAlertUpgradeMarshal failed, err: %v, event: %+v", err, event)
		return nil, err
	}

	// 补齐event中的字段
	event.Sname = stra.Name
	event.EndpointAlias = endpoint.Alias
	event.Category = stra.Category
	event.Priority = stra.Priority
	event.Nid = stra.Nid
	event.Users = string(users)
	event.Groups = string(groups)
	event.NodePath = nodePath
	event.NeedUpgrade = stra.NeedUpgrade
	event.AlertUpgrade = alertUpgrade
	err = model.SaveEvent(event)
	if err != nil {
		return event, err
	}

	if event.EventType == config.ALERT {
		eventCur := new(model.EventCur)
		if err = json.Unmarshal([]byte(reply[1]), eventCur); err != nil {
			logger.Errorf("unmarshal redis reply failed, err: %v, event: %+v", err, event)
		}

		eventCur.Sname = stra.Name
		eventCur.Category = stra.Category
		eventCur.Priority = stra.Priority
		eventCur.Nid = stra.Nid
		eventCur.Users = string(users)
		eventCur.Groups = string(groups)
		eventCur.NodePath = nodePath
		eventCur.NeedUpgrade = stra.NeedUpgrade
		eventCur.AlertUpgrade = alertUpgrade
		eventCur.EndpointAlias = endpoint.Alias
		eventCur.Status = 0
		eventCur.Claimants = "[]"
		err = model.SaveEventCur(eventCur)
		if err != nil {
			logger.Errorf("save event cur failed, err: %v, event: %+v", err, event)
			return event, nil
		}
	} else {
		err = model.EventCurDel(event.HashId)
		if err != nil {
			logger.Errorf("del event cur failed, err: %v, event: %v", err, event)
			return event, nil
		}
	}

	return event, nil
}
