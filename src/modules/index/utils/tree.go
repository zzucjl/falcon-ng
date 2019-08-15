package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"

	. "github.com/open-falcon/falcon-ng/src/modules/index/config"

	"github.com/toolkits/pkg/net/httplib"
)

type NidsReq struct {
	Host []string `json:"host"`
}

type LeafNidsReq struct {
	Ids []int64 `json:"ids"`
}

type NidResp struct {
	Dat map[string][]int64 `json:"dat"`
	Err string             `json:"err"`
}

func GetNidsFromTree(endpoint string) ([]int64, error) {
	var req NidsReq
	req.Host = []string{endpoint}

	res, err := GetDataFromTreeRetry("http://%s/api/node/query", req)
	if err != nil {
		return []int64{}, err
	}

	if len(res.Dat) > 0 {
		return res.Dat[endpoint], nil
	}

	return []int64{}, fmt.Errorf("not found nid by %s", endpoint)
}

func GetLeafNidsbyNid(nids []int64) ([]int64, error) {
	var req LeafNidsReq
	req.Ids = nids
	leafNids := []int64{}

	res, err := GetDataFromTreeRetry("http://%s/api/node/leafids", req)
	if err != nil {
		return []int64{}, err
	}
	for _, datas := range res.Dat {
		leafNids = append(leafNids, datas...)
	}

	return leafNids, nil
}

func GetDataFromTreeRetry(path string, req interface{}) (NidResp, error) {
	count := len(Config.Tree.Addrs)
	var resp NidResp
	var err error
	for i := 0; i < count; i++ {
		resp, err = GetDataFromTree(path, req)
		if err == nil {
			return resp, err
		}
	}
	return resp, err
}

func GetDataFromTree(path string, req interface{}) (NidResp, error) {
	var res NidResp
	i := rand.Intn(len(Config.Tree.Addrs))
	addr := Config.Tree.Addrs[i]

	url := fmt.Sprintf(path, addr)
	resp, err := httplib.PostJSON(url, Config.Tree.Timeout, req)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(resp, &res)
	return res, err
}
