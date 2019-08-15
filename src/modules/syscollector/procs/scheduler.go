package procs

import (
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/nux"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/funcs"
)

type ProcScheduler struct {
	Ticker *time.Ticker
	Proc   *model.ProcCollect
	Quit   chan struct{}
}

func NewProcScheduler(p *model.ProcCollect) *ProcScheduler {
	scheduler := ProcScheduler{Proc: p}
	scheduler.Ticker = time.NewTicker(time.Duration(p.Step) * time.Second)
	scheduler.Quit = make(chan struct{})
	return &scheduler
}

func (this *ProcScheduler) Schedule() {
	go func() {
		for {
			select {
			case <-this.Ticker.C:
				ProcCollect(this.Proc)
			case <-this.Quit:
				this.Ticker.Stop()
				return
			}
		}
	}()
}

func (this *ProcScheduler) Stop() {
	close(this.Quit)
}

func ProcCollect(p *model.ProcCollect) {
	ps, err := nux.AllProcs()
	if err != nil {
		logger.Error(err)
		return
	}

	pslen := len(ps)
	cnt := 0
	for i := 0; i < pslen; i++ {
		if isProc(ps[i], p.CollectMethod, p.Target) {
			cnt++
		}
	}

	item := funcs.GaugeValue("proc.num", cnt, p.Tags)
	item.Step = int64(p.Step)
	item.Timestamp = time.Now().Unix()
	item.Endpoint = config.Hostname

	funcs.Push([]*dataobj.MetricValue{item})
}

func isProc(p *nux.Proc, method, target string) bool {
	if method == "name" && target == p.Name {
		return true
	} else if (method == "cmdline" || method == "cmd") && strings.Contains(p.Cmdline, target) {
		return true
	}
	return false
}
