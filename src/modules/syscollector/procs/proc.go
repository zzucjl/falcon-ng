package procs

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
	Procs              = make(map[string]*model.ProcCollect)
	ProcsWithScheduler = make(map[string]*ProcScheduler)
)

func DelNoPorcCollect(newCollect map[string]*model.ProcCollect) {
	for currKey, currProc := range Procs {
		newProc, ok := newCollect[currKey]
		if !ok || currProc.LastUpdated != newProc.LastUpdated {
			deleteProc(currKey)
		}
	}
}

func AddNewPorcCollect(newCollect map[string]*model.ProcCollect) {
	for target, newProc := range newCollect {
		if _, ok := Procs[target]; ok && newProc.LastUpdated == Procs[target].LastUpdated {
			continue
		}

		Procs[target] = newProc
		sch := NewProcScheduler(newProc)
		ProcsWithScheduler[target] = sch
		sch.Schedule()
	}
}

func deleteProc(key string) {
	v, ok := ProcsWithScheduler[key]
	if ok {
		v.Stop()
		delete(ProcsWithScheduler, key)
	}
	delete(Procs, key)
}

func NewProcCollect(method, name, tags string, step int) *model.ProcCollect {
	return &model.ProcCollect{
		CollectType:   "proc",
		CollectMethod: method,
		Target:        name,
		Step:          step,
		Tags:          tags,
	}
}

func ListProcs() map[string]*model.ProcCollect {
	procs := make(map[string]*model.ProcCollect)

	if config.Config.Collector.SyncCollect {
		procs = config.Collect.GetProcs()
		for _, p := range procs {
			tagsMap := utils.DictedTagstring(p.Tags)
			tagsMap["target"] = p.Target
			p.Tags = utils.SortedTags(tagsMap)
		}
	}

	files, err := file.FilesUnder(config.Config.ProcPath)
	if err != nil {
		logger.Error(err)
		return procs
	}

	//扫描文件采集配置
	for _, f := range files {
		method, name, step, err := parseName(f)
		if err != nil {
			logger.Warning(err)
			continue
		}

		service, err := file.ToTrimString(config.Config.ProcPath + "/" + f)
		if err != nil {
			logger.Warning(err)
			continue
		}

		tags := fmt.Sprintf("target=%s,service=%s", name, service)
		p := NewProcCollect(method, name, tags, step)
		procs[p.Name] = p
	}

	return procs
}

func parseName(fname string) (method string, name string, step int, err error) {
	arr := strings.Split(fname, "_")
	if len(arr) < 3 {
		err = fmt.Errorf("name is illegal %s, split _ < 3", fname)
		return
	}

	step, err = strconv.Atoi(arr[0])
	if err != nil {
		err = fmt.Errorf("name is illegal %s %v", fname, err)
		return
	}

	method = arr[1]

	name = strings.Join(arr[2:len(arr)], "_")
	return
}
