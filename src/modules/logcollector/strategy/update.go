package strategy

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/schema"
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/utils"

	"github.com/parnurzeal/gorequest"
)

const PATTERN_EXCLUDE_PARTITION = "```EXCLUDE```"

func Update() error {
	markTms := time.Now().Unix()
	logger.Infof("[%d]Update Strategy start", markTms)
	strategys, err := GetCollects()
	if err != nil {
		logger.Errorf("[%d]Get my Strategy error ! [msg:%v]", markTms, err)
		return err
	} else {
		logger.Infof("[%d]Get my Strategy success, num : [%d]", markTms, len(strategys))
	}
	parsePattern(strategys)
	updateRegs(strategys)

	err = UpdateGlobalStrategy(strategys)
	if err != nil {
		logger.Errorf("[%d]Update Strategy cache error ! [msg:%v]", err)
		return err
	}
	logger.Infof("[%d]Update Strategy end", markTms)
	return nil
}

//parsePattern 规范化处理
func parsePattern(strategys []*schema.Strategy) {
	for _, st := range strategys {
		patList := strings.Split(st.Pattern, PATTERN_EXCLUDE_PARTITION)

		if len(patList) == 1 {
			st.Pattern = strings.TrimSpace(st.Pattern)
			continue
		} else if len(patList) >= 2 {
			st.Pattern = strings.TrimSpace(patList[0])
			st.Exclude = strings.TrimSpace(patList[1])
			continue
		} else {
			logger.Errorf("bad pattern to parse : [%s]", st.Pattern)
		}
	}
}

func updateRegs(strategys []*schema.Strategy) {
	for _, st := range strategys {
		st.TagRegs = make(map[string]*regexp.Regexp, 0)
		st.ParseSucc = false

		//更新时间正则
		pat, _ := utils.GetPatAndTimeFormat(st.TimeFormat)
		reg, err := regexp.Compile(pat)
		if err != nil {
			logger.Errorf("compile time regexp failed:[sid:%d][format:%s][pat:%s][err:%v]", st.ID, st.TimeFormat, pat, err)
			continue
		}
		st.TimeReg = reg //拿到时间正则

		if len(st.Pattern) == 0 && len(st.Exclude) == 0 {
			logger.Errorf("pattern and exclude are all empty, sid:[%d]", st.ID)
			continue
		}

		//更新pattern
		if len(st.Pattern) != 0 {
			reg, err = regexp.Compile(st.Pattern)
			if err != nil {
				logger.Errorf("compile pattern regexp failed:[sid:%d][pat:%s][err:%v]", st.ID, st.Pattern, err)
				continue
			}
			st.PatternReg = reg
		}

		//更新exclude
		if len(st.Exclude) != 0 {
			reg, err = regexp.Compile(st.Exclude)
			if err != nil {
				logger.Errorf("compile exclude regexp failed:[sid:%d][pat:%s][err:%v]", st.ID, st.Exclude, err)
				continue
			}
			st.ExcludeReg = reg
		}

		//更新tags
		for tagk, tagv := range st.Tags {
			reg, err = regexp.Compile(tagv)
			if err != nil {
				logger.Errorf("compile tag failed:[sid:%d][pat:%s][err:%v]", st.ID, st.Exclude, err)
				continue
			}
			st.TagRegs[tagk] = reg
		}
		st.ParseSucc = true
	}
}

type CollectResp struct {
	Dat model.Collect `json:"dat"`
	Err string        `json:"err"`
}

func GetCollects() ([]*schema.Strategy, error) {
	var res CollectResp
	var err error
	var stras []*schema.Strategy
	if config.Config.Strategy.SyncCollect {
		i := rand.Intn(len(config.Config.Strategy.ConfigAddrs))
		addr := config.Config.Strategy.ConfigAddrs[i]

		url := fmt.Sprintf("http://%s/api/collect/%s", addr, config.Hostname)

		_, _, errs := gorequest.New().Timeout(time.Duration(config.Config.Strategy.Timeout) * time.Second).Get(url).EndStruct(&res)
		if len(errs) != 0 {
			err = fmt.Errorf("get collects from remote failed, error:%v", errs)
			logger.Info("Timeout ", config.Config.Strategy.Timeout)
			return stras, err
		} else {
			strasMap := res.Dat.GetLogConfig()
			for _, s := range strasMap {
				stra := DeepCopyStrategy(s)
				stras = append(stras, stra)
			}
		}
	}

	//从文件中读取
	stras = append(stras, GetCollectFromFile()...)

	return stras, err
}

func GetCollectFromFile() []*schema.Strategy {
	logger.Info("get collects from local file")
	var stras []*schema.Strategy

	files, err := file.FilesUnder(config.Config.Strategy.FilePath)
	if err != nil {
		logger.Error(err)
		return stras
	}

	//扫描文件采集配置
	for _, f := range files {
		err := checkName(f)
		if err != nil {
			logger.Warningf("read file name err:%s %v", f, err)
			continue
		}

		stra := schema.Strategy{}

		b, err := file.ToBytes(config.Config.Strategy.FilePath + "/" + f)
		if err != nil {
			logger.Warningf("read file name err:%s %v", f, err)
			continue
		}

		err = json.Unmarshal(b, &stra)
		if err != nil {
			logger.Warningf("read file name err:%s %v", f, err)
			continue
		}

		//todo 配置校验

		stra.ID = genStraID(stra.Name, string(b))
		stras = append(stras, &stra)
	}

	return stras
}

func genStraID(name, body string) int64 {
	var id int64
	all := name + body
	if len(all) < 1 {
		return id
	}

	id = int64(all[0])

	for i := 1; i < len(all); i++ {
		id += int64(all[i])
		id += int64(all[i] - all[i-1])
	}

	id += 1000000 //避免与web端配置的id冲突
	return id
}

func checkName(f string) (err error) {
	if !strings.Contains(f, "log.") {
		err = fmt.Errorf("name illege not contain log. %s", f)
		return
	}

	arr := strings.Split(f, ".")
	if len(arr) < 3 {
		err = fmt.Errorf("name illege %s len:%d len < 3 ", f, len(arr))
		return
	}

	if arr[len(arr)-1] != "json" {
		err = fmt.Errorf("name illege %s not json file", f)
		return
	}

	return
}

func DeepCopyStrategy(p *model.LogCollect) *schema.Strategy {
	s := schema.Strategy{}
	s.ID = p.Id
	s.Name = p.Name
	s.FilePath = p.FilePath
	s.TimeFormat = p.TimeFormat
	s.Pattern = p.Pattern
	s.MeasurementType = p.CollectType
	s.Interval = int64(p.Step)
	s.Tags = DeepCopyStringMap(p.Tags)
	s.Func = p.Func
	s.Degree = int64(p.Degree)
	s.Unit = p.Unit
	s.Comment = p.Comment
	s.Creator = p.Creator
	s.SrvUpdated = p.LastUpdated.String()
	s.LocalUpdated = p.LocalUpdated

	return &s
}

func DeepCopyStringMap(p map[string]string) map[string]string {
	r := make(map[string]string, len(p))
	for k, v := range p {
		r[k] = v
	}
	return r
}
