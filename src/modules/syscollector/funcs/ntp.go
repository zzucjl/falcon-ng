package funcs

import (
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/nux"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
)

var ntpServer string

func NtpOffsetMetrics() (L []*dataobj.MetricValue) {
	ntpServers := config.Config.NtpServers
	if len(ntpServers) <= 0 {
		return
	}

	for idx, server := range ntpServers {
		if ntpServer == "" {
			ntpServer = server
		}
		orgTime := time.Now()
		logger.Debug("ntp: use server, ", ntpServer)
		logger.Debug("ntp: client send time, ", orgTime)
		serverReciveTime, serverTransmitTime, err := nux.NtpTwoTime(ntpServer)
		if err != nil {
			logger.Warning("ntp: get err", ntpServer, err)
			ntpServer = ""
			time.Sleep(time.Second * time.Duration(idx+1))
			continue
		} else {
			ntpServer = server //找一台正常的ntp一直使用
		}
		dstTime := time.Now()
		// 算法见https://en.wikipedia.org/wiki/Network_Time_Protocol
		duration := ((serverReciveTime.UnixNano() - orgTime.UnixNano()) + (serverTransmitTime.UnixNano() - dstTime.UnixNano())) / 2
		logger.Debug("ntp: server receive time, ", serverReciveTime)
		logger.Debug("ntp: server reply time, ", serverTransmitTime)
		logger.Debug("ntp: client receive time, ", dstTime)

		delta := duration / 1e6 // 转换成 ms
		L = append(L, GaugeValue("sys.ntp.offset.ms", delta))
		//one ntp server's response is enough

		return
	}

	//keep silence when no config ntp server
	if len(ntpServers) > 0 {
		logger.Error("sys.ntp.offset error. all ntp servers response failed.")
	}
	return
}
