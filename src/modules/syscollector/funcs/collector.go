package funcs

import (
	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
)

func CollectorMetrics() []*dataobj.MetricValue {
	return []*dataobj.MetricValue{
		GaugeValue("proc.agent.alive", 1),
		GaugeValue("proc.agent.version", config.Version),
	}
}
