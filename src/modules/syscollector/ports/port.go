package ports

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/utils"
)

var (
	Ports              = make(map[int]*model.PortCollect)
	PortsWithScheduler = make(map[int]*PortScheduler)
)

func DelNoPortCollect(newCollect map[int]*model.PortCollect) {
	for currKey, currPort := range Ports {
		newPort, ok := newCollect[currKey]
		if !ok || currPort.LastUpdated != newPort.LastUpdated {
			deletePort(currKey)
		}
	}
}

func AddNewPortCollect(newCollect map[int]*model.PortCollect) {
	for target, newPort := range newCollect {
		if _, ok := Ports[target]; ok && newPort.LastUpdated == Ports[target].LastUpdated {
			continue
		}

		Ports[target] = newPort
		sch := NewPortScheduler(newPort)
		PortsWithScheduler[target] = sch
		sch.Schedule()
	}
}

func deletePort(key int) {
	v, ok := PortsWithScheduler[key]
	if ok {
		v.Stop()
		delete(PortsWithScheduler, key)
	}
	delete(Ports, key)
}

func NewPortCollect(port, step int, tags string) *model.PortCollect {
	return &model.PortCollect{
		CollectType: "port",
		Port:        port,
		Step:        step,
		Tags:        tags,
	}
}

func ListPorts() map[int]*model.PortCollect {
	ports := make(map[int]*model.PortCollect)

	if config.Config.Collector.SyncCollect {
		ports = config.Collect.GetPorts()
		for _, p := range ports {
			tagsMap := utils.DictedTagstring(p.Tags)
			tagsMap["port"] = strconv.Itoa(p.Port)

			p.Tags = utils.SortedTags(tagsMap)
		}
	}

	files, err := file.FilesUnder(config.Config.PortPath)
	if err != nil {
		logger.Error(err)
		return ports
	}
	//扫描文件采集配置
	for _, f := range files {
		port, step, err := parseName(f)
		if err != nil {
			logger.Warning(err)
			continue
		}

		service, err := file.ToTrimString(config.Config.PortPath + "/" + f)
		if err != nil {
			logger.Warning(err)
			continue
		}

		tags := fmt.Sprintf("port=%s,service=%s", strconv.Itoa(port), service)
		p := NewPortCollect(port, step, tags)
		ports[p.Port] = p
	}

	return ports
}

func parseName(name string) (port, step int, err error) {
	arr := strings.Split(name, "_")
	if len(arr) < 2 {
		err = fmt.Errorf("name is illegal %s, split _ < 2", name)

		return
	}

	step, err = strconv.Atoi(arr[0])
	if err != nil {
		err = fmt.Errorf("name is illegal %s %v", name, err)
		return
	}

	port, err = strconv.Atoi(arr[1])
	if err != nil {
		err = fmt.Errorf("name is illegal %s %v", name, err)
		return
	}
	return
}
