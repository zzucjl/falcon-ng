package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"reflect"
	"time"

	"github.com/toolkits/pkg/concurrent/semaphore"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/ugorji/go/codec"
)

func main() {

	log.Println()
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	conn, err := net.DialTimeout("tcp", "10.86.76.13:8041", time.Millisecond*300)
	if err != nil {
		log.Fatalln(err)
	}

	var bufconn = struct { // bufconn here is a buffered io.ReadWriteCloser
		io.Closer
		*bufio.Reader
		*bufio.Writer
	}{conn, bufio.NewReader(conn), bufio.NewWriter(conn)}

	rpcCodec := codec.MsgpackSpecRpc.ClientCodec(bufconn, &mh)
	client := rpc.NewClientWithCodec(rpcCodec)
	sema := semaphore.NewSemaphore(10)
	log.Println("begin...")
	for {
		for i := 0; i < 10; i++ {
			sema.Acquire()
			go func() {
				defer sema.Release()
				query := MakeQuery()
				var reply *dataobj.QueryDataResp
				err = client.Call("Transfer.Query", query, &reply)
				if err != nil {
					log.Println(err)
				} else {
					for _, d := range reply.Data {
						log.Println("rpc succ, reply: ", d)
					}
				}
			}()
		}
		time.Sleep(1 * time.Second)
		log.Println("query...")
	}

	log.Println("end...")

	client.Close()
}

func MakeQuery() []dataobj.QueryData {
	counter := fmt.Sprintf("metric.%d/service=tsdb.%d", rand.Intn(1000), rand.Intn(100))
	req := dataobj.QueryData{
		Start:      time.Now().Unix() - 300,
		End:        time.Now().Unix(),
		Step:       10,
		ConsolFunc: "AVERAGE",
		Endpoints:  []string{"endpoint.0"},
		Counters:   []string{counter},
		DsType:     "GAUGE",
	}

	return []dataobj.QueryData{req}
}
