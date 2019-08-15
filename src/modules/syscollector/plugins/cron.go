package plugins

import (
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
)

func Detect() {
	detect()
	go loopDetect()
}

func loopDetect() {
	for {
		time.Sleep(time.Second * 10)
		detect()
	}
}

func detect() {
	ps := ListPlugins(config.Config.Plugin)
	DelNoUsePlugins(ps)
	AddNewPlugins(ps)
}
