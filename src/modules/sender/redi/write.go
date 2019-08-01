package redi

import (
	"strings"

	"github.com/json-iterator/go"
	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

func lpush(queue, message string) {
	rc := RedisConnPool.Get()
	defer rc.Close()
	_, err := rc.Do("LPUSH", queue, message)
	if err != nil {
		logger.Error("LPUSH redis", queue, "fail:", err, "message:", message)
	}
}

func Write(data *dataobj.Notify, notifyType string) {
	if data == nil {
		return
	}

	data.Tos = removeEmptyString(data.Tos)

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	bs, err := json.Marshal(data)
	if err != nil {
		logger.Error("marshal mail failed, dat: %+v, err: %v", data, err)
		return
	}

	queue := ChoiceQueue(notifyType)
	lpush(queue, string(bs))
	logger.Debugf("write mail to queue, mail:%v, queue:%s", data, queue)
}

func removeEmptyString(s []string) []string {
	ss := []string{}
	for i := 0; i < len(s); i++ {
		if strings.TrimSpace(s[i]) == "" {
			continue
		}

		ss = append(ss, s[i])
	}

	return ss
}
