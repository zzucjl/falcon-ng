package funcs

import (
	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
)

type FuncsAndInterval struct {
	Fs       []func() []*dataobj.MetricValue
	Interval int
}

var Mappers []FuncsAndInterval

func BuildMappers() {
	interval := config.Config.Transfer.Interval
	Mappers = []FuncsAndInterval{
		{
			Fs: []func() []*dataobj.MetricValue{
				CollectorMetrics,
				CpuMetrics,
				MemMetrics,
				NetMetrics,
				LoadAvgMetrics,
				IOStatsMetrics,
				NfMetrics,
				FsKernelMetrics,
				FsRWMetrics,
				ProcsNumMetrics,
				EntityNumMetrics,
				NtpOffsetMetrics,
				SocketStatSummaryMetrics,
			},
			Interval: interval,
		},
		{
			Fs: []func() []*dataobj.MetricValue{
				DeviceMetrics,
			},
			Interval: interval,
		},
	}
}
