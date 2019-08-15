package rpc

import (
	"fmt"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/backend"
	. "github.com/open-falcon/falcon-ng/src/modules/transfer/config"
)

type TransferResp struct {
	Msg     string
	Total   int
	Invalid int
	Latency int64
}

func (t *TransferResp) String() string {
	s := fmt.Sprintf("TransferResp total=%d, err_invalid=%d, latency=%dms",
		t.Total, t.Invalid, t.Latency)
	if t.Msg != "" {
		s = fmt.Sprintf("%s, msg=%s", s, t.Msg)
	}
	return s
}

func (this *Transfer) Ping(args string, reply *string) error {
	*reply = args
	return nil
}

func (this *Transfer) Push(args []*dataobj.MetricValue, reply *TransferResp) error {
	start := time.Now()
	reply.Invalid = 0

	items := []*dataobj.MetricValue{}
	for _, v := range args {
		logger.Debug("->recv: ", v)
		err := v.CheckValidity()
		if err != nil {
			logger.Warningf("item is illegal item:%s err:%v", v, err)
			reply.Invalid += 1
			reply.Msg += fmt.Sprintf("%v\n", err)
			continue
		}
		logger.Debug("->check ok: ", v)

		items = append(items, v)
	}

	if Config.Tsdb.Enabled {
		backend.Push2TsdbSendQueue(items)
	}
	if reply.Invalid == 0 {
		reply.Msg = "ok"
	}

	reply.Total = len(args)
	reply.Latency = (time.Now().UnixNano() - start.UnixNano()) / 1000000
	return nil
}
