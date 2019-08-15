package rpc

import (
	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/backend"
)

func (this *Transfer) Query(args []dataobj.QueryData, reply *dataobj.QueryDataResp) error {
	//start := time.Now()
	reply.Data = backend.FetchData(args)
	return nil
}
