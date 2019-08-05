package buffer

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

func Test_RingHistory(t *testing.T) {
	h := NewRingHistory(18, 0, 10)

	points := make([]*dataobj.RRDData, 30)
	for i := 0; i < 30; i++ {
		points[i] = &dataobj.RRDData{Timestamp: int64(i*10 + 1), Value: dataobj.JsonFloat(i)}
	}

	for i := 1; i < 19; i++ {
		h.Write(points[i : i+10])
		debug(h)
	}

	ps := h.Read(120, 200)
	for i := range ps {
		fmt.Println(ps[i].Timestamp, ps[i].Value)
	}
}

func Benchmark_RingWrite(b *testing.B) {
	b.StopTimer()
	h := NewRingHistory(18, 0, 10)
	points := make([]*dataobj.RRDData, 30)
	for i := 0; i < 30; i++ {
		points[i] = &dataobj.RRDData{Timestamp: int64(i), Value: dataobj.JsonFloat(i)}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.Write(points[1:11])
	}
}

func Test_ChainHistory(t *testing.T) {
	h := NewChainHistory(18)
	points := make([]*dataobj.RRDData, 30)
	for i := 0; i < 30; i++ {
		points[i] = &dataobj.RRDData{Timestamp: int64(i*10 + 1), Value: dataobj.JsonFloat(i)}
	}

	for i := 1; i < 19; i++ {
		h.Write(points[i : i+2])
		points := h.Dump()
		for j := range points {
			fmt.Printf("%d\t", points[j].Timestamp)
		}
		fmt.Println(len(points))
	}
}

func Test_ChainDisorder(t *testing.T) {
	h := NewChainHistory(3)
	var points []*dataobj.RRDData

	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054480,
		Value:     2,
	})
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054470,
		Value:     2,
	})
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054460,
		Value:     2,
	})
	h.Write(points)
	hpoints := h.Dump()
	bytes, _ := json.Marshal(hpoints)
	fmt.Println(string(bytes))

	points = points[:0]
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054490,
		Value:     2,
	})
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054480,
		Value:     2,
	})
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054470,
		Value:     2,
	})
	h.Write(points)
	hpoints = h.Dump()
	bytes, _ = json.Marshal(hpoints)
	fmt.Println(string(bytes))

	points = points[:0]
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054480,
		Value:     2,
	})
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054500,
		Value:     2,
	})
	points = append(points, &dataobj.RRDData{
		Timestamp: 1553054490,
		Value:     2,
	})
	h.Write(points)
	hpoints = h.Dump()
	bytes, _ = json.Marshal(hpoints)
	fmt.Println(string(bytes))
}

func Test_ChainCleanup(t *testing.T) {
	h := NewChainHistory(18)
	points := make([]*dataobj.RRDData, 1)

	for i := 1; i < 3; i++ {
		points[0] = &dataobj.RRDData{Timestamp: int64(i%3 + 1), Value: dataobj.JsonFloat(1)}
		h.Write(points)
		points := h.Dump()
		for j := range points {
			fmt.Printf("%d\t", points[j].Timestamp)
		}
		fmt.Println("")
		h.Cleanup()
		points = h.Dump()
		for j := range points {
			fmt.Printf("%d\t", points[j].Timestamp)
		}
		fmt.Println("")
	}
}

func Benchmark_ChainWrite(b *testing.B) {
	b.StopTimer()
	h := NewChainHistory(18)
	points := make([]*dataobj.RRDData, 30)
	for i := 0; i < 30; i++ {
		points[i] = &dataobj.RRDData{Timestamp: int64(i), Value: dataobj.JsonFloat(i)}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.Write(points[1:11])
	}
}

func debug(h *RingHistory) {
	fmt.Printf("start:%d, end:%d, size:%d\n", h.start, h.end, len(h.data))
	fmt.Printf("%v\n", h.data)
}
