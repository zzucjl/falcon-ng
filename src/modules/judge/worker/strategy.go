package worker

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"

	"github.com/parnurzeal/gorequest"
)

func GetStrategyFromRemote(opts StrategyConfigOption) ([]*schema.Strategy, error) {
	if len(opts.Addrs) == 0 {
		return nil, errors.New("empty config addr")
	}
	client := gorequest.New().
		Timeout(time.Duration(opts.Timeout) * time.Millisecond)
	stras := struct {
		Data []*model.Stra `json:"dat"`
		Err  string        `json:"err"`
	}{}
	perm := rand.Perm(len(opts.Addrs))
	for i := range perm {
		url := fmt.Sprintf("http://%s"+opts.PartitionApi, opts.Addrs[perm[i]], identity)
		resp, _, errs := client.
			Get(url).
			EndStruct(&stras)

		if len(errs) != 0 {
			logger.Debugf(0, "get strategy from remote failed, error:%v", errs)
			continue
		}

		if resp == nil {
			logger.Debug(0, "get strategy from remote failed, error: nil resp")
			continue
		}
		if resp.StatusCode != http.StatusOK {
			logger.Debugf(0, "get strategy from remote failed, code:%d, error:%v",
				resp.StatusCode, errs)
			continue
		}

		if len(stras.Data) > 0 {
			break
		}
	}

	ret := make([]*schema.Strategy, 0)
	for i := range stras.Data {
		s, err := parseStrategyFromRemote(stras.Data[i])
		if err != nil {
			bytes, _ := json.Marshal(stras.Data[i])
			logger.Warningf(0, "parse strategy failed, error:%v, strategy:%s", err, string(bytes))
			continue
		}
		ret = append(ret, s)
	}

	return ret, nil
}

func parseStrategyFromRemote(s *model.Stra) (*schema.Strategy, error) {
	if s == nil {
		return nil, errors.New("nil pointer")
	}
	if len(s.LeafNids) == 0 {
		return nil, errors.New("empty leaf nid")
	}

	if s.Exprs == nil {
		return nil, errors.New("empty expression")
	}

	if len(s.Exprs) == 0 {
		return nil, errors.New("empty expression")
	}

	judgements, err := parseJudgement(s, s.Exprs, s.Tags)
	if err != nil {
		return nil, err
	}

	alert, err := parseAlert(s)
	if err != nil {
		return nil, err
	}

	endpMap := make(map[string]struct{})
	for i := range s.Endpoints {
		endpMap[s.Endpoints[i]] = struct{}{}
	}
	endps := make([]string, len(endpMap))
	i := 0
	for endp, _ := range endpMap {
		endps[i] = endp
		i++
	}

	return &schema.Strategy{
		ID:         s.Id,
		Name:       s.Name,
		Priority:   s.Priority,
		Category:   s.Category,
		Operator:   schema.LOGIC_OPERATOR_AND,
		WindowSize: 180, // 硬编码, 数据断点时, 最多等待的周期数
		Endpoints:  endps,
		Partition:  "/event/p" + strconv.Itoa(s.Priority), // event的队列名
		Judgements: judgements,
		Alert:      alert,
		Updated:    s.LastUpdated.Unix(),
	}, nil
}

func parseJudgement(s *model.Stra, exps []model.Exp,
	tags []model.Tag) ([]schema.StrategyJudgement, error) {
	if s.AlertDur <= 0 {
		return nil, errors.New("alert duration empty")
	}

	var filters []schema.StrategyTagFilter
	if len(tags) > 0 {
		for i := range tags {
			filters = append(filters, schema.StrategyTagFilter{
				Tagk:     tags[i].Tkey,
				Operator: tags[i].Topt,
				Tagv:     tags[i].Tval,
			})
		}
	}
	var judgements []schema.StrategyJudgement
	for i := range exps {
		var (
			trigger string // 转化为judge可以识别的trigger和params
			params  []string
		)

		switch exps[i].Func {
		case "all":
			trigger = schema.TRIGGER_DURATION_STAT
			params = append(params, strconv.Itoa(s.AlertDur))
			params = append(params, schema.MATH_OPERATOR_OBO)
		case "max":
			trigger = schema.TRIGGER_DURATION_STAT
			params = append(params, strconv.Itoa(s.AlertDur))
			params = append(params, schema.MATH_OPERATOR_MAX)
		case "min":
			trigger = schema.TRIGGER_DURATION_STAT
			params = append(params, strconv.Itoa(s.AlertDur))
			params = append(params, schema.MATH_OPERATOR_MIN)
		case "sum":
			trigger = schema.TRIGGER_DURATION_STAT
			params = append(params, strconv.Itoa(s.AlertDur))
			params = append(params, schema.MATH_OPERATOR_SUM)
		case "avg":
			trigger = schema.TRIGGER_DURATION_STAT
			params = append(params, strconv.Itoa(s.AlertDur))
			params = append(params, schema.MATH_OPERATOR_AVG)
		case "happen":
			trigger = schema.TRIGGER_DURATION_HAPPEN
			if len(exps[i].Params) != 1 {
				return nil, errors.New("func happen param illegal")
			}
			params = append(params, strconv.Itoa(s.AlertDur))
			params = append(params, strconv.Itoa(exps[i].Params[0]))
		case "nodata":
			trigger = schema.TRIGGER_NODATA
			params = append(params, strconv.Itoa(s.AlertDur))
		default:
			return nil, errors.New("func not support")
		}
		execution := schema.StrategyExecution{
			EffectiveDay:   []int{0, 1, 2, 3, 4, 5, 6},
			EffectiveStart: 0,
			EffectiveEnd:   1439,                      // 默认设置为一直有效
			Operator:       schema.LOGIC_OPERATOR_AND, // 默认是 "and"
			Expressions: []schema.StrategyExpression{
				schema.StrategyExpression{
					Func:     trigger,
					Params:   params,
					Operator: schema.LOGIC_OPERATOR_AND, // 默认是 "and"
					Thresholds: []schema.StrategyThreshold{
						schema.StrategyThreshold{
							Operator: exps[i].Eopt, Threshold: float64(exps[i].Threshold)},
					},
				},
			},
		}
		judgements = append(judgements, schema.StrategyJudgement{
			Metric:    exps[i].Metric,
			Tags:      filters,
			Execution: execution,
		})
	}
	return judgements, nil
}

func parseAlert(s *model.Stra) (schema.StrategyAlert, error) {
	return schema.StrategyAlert{
		AlertCountThreshold:      1, // 统一设置为1, alertDuration 字段用于算子参数
		RecoverDurationThreshold: s.RecoveryDur,
		LimitCountThreshold:      1,
		LimitDurationThreshold:   300,
	}, nil
}
