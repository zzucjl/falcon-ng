package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/cache"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/config"
)

func SyncMaskconfLoop() {
	interval := config.GetCfgYml().Interval
	for {
		SyncMaskconf()
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func SyncMaskconf() error {
	err := model.CleanExpireMask(time.Now().Unix())
	if err != nil {
		logger.Errorf("clean expire mask fail, err: %v", err)
		return err
	}

	mcs, err := model.MaskconfGetAll()
	if err != nil {
		logger.Errorf("get maskconf fail, err: %v", err)
		return err
	}

	// key: metric#endpoint
	// value: tags
	maskMap := make(map[string][]string)
	for i := 0; i < len(mcs); i++ {
		err := mcs[i].FillEndpoints()
		if err != nil {
			logger.Errorf("%v fill endpoints fail, err: %v", mcs[i], err)
			return err
		}

		for j := 0; j < len(mcs[i].Endpoints); j++ {
			key := mcs[i].Metric + "#" + mcs[i].Endpoints[j]
			maskMap[key] = append(maskMap[key], tagsFormat(mcs[i].Tags)...)
		}
	}

	cache.MaskCache.SetAll(maskMap)

	return nil
}

func IsMaskEvent(event *model.Event) bool {
	detail, err := event.GetEventDetail()
	if err != nil {
		logger.Errorf("get event detail failed, err: %v", err)
		return false
	}

	for i := 0; i < len(detail); i++ {
		eventMetric := detail[i].Metric
		eventTagsList := []string{}

		for k, v := range detail[i].Tags {
			eventTagsList = append(eventTagsList, fmt.Sprintf("%s=%s", strings.TrimSpace(k), strings.TrimSpace(v)))
		}
		key := eventMetric + "#" + event.Endpoint
		maskTagsList, exists := cache.MaskCache.GetByKey(key)
		if !exists {
			continue
		}

		for i := 0; i < len(maskTagsList); i++ {
			tagsList := strings.Split(maskTagsList[i], ",")
			if inList("", tagsList) {
				return true
			}

			if listContains(tagsList, eventTagsList) {
				return true
			}
		}
	}

	return false
}

// tags: key1=value1,value2;key2=value3,value4  => ["key1=value1,key2=value3", "key1=value1,key2=value4", "key1=value2,key2=value3", "key1=value2,key2=value4"]
func tagsFormat(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return []string{""}
	}

	arr := make(map[string][]string)
	tagsList := strings.Split(tags, ";")
	for i := 0; i < len(tagsList); i++ {
		kv := strings.Split(tagsList[i], "=")
		key := kv[0]
		valueList := strings.Split(kv[1], ",")

		l := []string{}
		if _, has := arr[key]; has {
			l = arr[key]
		}

		for j := 0; j < len(valueList); j++ {
			l = append(l, fmt.Sprintf("%s=%s", strings.TrimSpace(key), strings.TrimSpace(valueList[j])))
		}

		arr[key] = l
	}

	if len(arr) == 1 {
		for _, val := range arr {
			return val
		}
	}

	t := []string{}
	i := 1
	for _, val := range arr {
		if i == 1 {
			t = val
			i++
			continue
		}

		m := []string{}
		for j := 0; j < len(val); j++ {
			for k := 0; k < len(t); k++ {
				m = append(m, fmt.Sprintf("%s,%s", t[k], val[j]))
			}
		}
		t = m
	}

	return t
}

// 用来判断blist是否包含slist
func listContains(slist, blist []string) bool {
	for i := 0; i < len(slist); i++ {
		if !inList(slist[i], blist) {
			return false
		}
	}

	return true
}

func inList(v string, lst []string) bool {
	for i := 0; i < len(lst); i++ {
		if lst[i] == v {
			return true
		}
	}

	return false
}
