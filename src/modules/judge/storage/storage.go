package storage

import (
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

type Storage interface {
	Query(ID uint32, stime, etime int64, span int) ([]*dataobj.RRDData, error)
	Index(req *IndexRequest) ([]Counter, error)
	GenerateAndSet(s *series.Series, bufferSize int, spans []int) uint32
	Get(ID uint32) (*series.Series, bool)
	Cleanup()
}

type IndexRequest struct {
	Endpoints []string            `json:"endpoints"`
	Metric    string              `json:"metric"`
	Include   map[string][]string `json:"include"`
	Exclude   map[string][]string `json:"exclude"`
}

type Counter struct {
	Counter string `json:"counter"`
	Step    int    `json:"step"`
	Dstype  string `json:"dstype`
}
