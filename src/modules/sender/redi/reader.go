package redi

import (
	"github.com/garyburd/redigo/redis"
	"github.com/json-iterator/go"
	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

func Pop(count int, notifyType string) []*dataobj.Notify {
	ret := []*dataobj.Notify{}
	queue := ChoiceQueue(notifyType)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	rc := RedisConnPool.Get()
	defer rc.Close()

	for i := 0; i < count; i++ {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				logger.Errorf("rpop queue:%s failed, err: %v", queue, err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var mail dataobj.Notify
		err = json.Unmarshal([]byte(reply), &mail)
		if err != nil {
			logger.Errorf("unmarshal mail failed, err: %v, redis reply: %v", err, reply)
			continue
		}

		ret = append(ret, &mail)
	}

	return ret
}
