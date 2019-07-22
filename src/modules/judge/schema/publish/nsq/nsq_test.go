package nsq

import (
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/stretchr/testify/assert"
)

func Test_Post(t *testing.T) {
	opts := publish.NsqPublisherOption{
		Addrs:                []string{"http://127.0.0.1:8080/put"},
		CallTimeout:          1000,
		BufferSize:           10,
		BufferEnqueueTimeout: 100,
	}
	go mockNsq()

	nsqp, _ := NewNsqPublisher(opts)

	event := &schema.Event{
		Sid:       1,
		EventType: "alert",
		Hashid:    uint64(111111),
		Etime:     1234556,
		History: []schema.History{schema.History{
			Key:    "cpu.idle/host=mock",
			Metric: "cpu.idle",
			Counter: map[string]string{
				"host": "mock",
			},
			Points: []*dataobj.RRDData{&dataobj.RRDData{
				Timestamp: 12345678,
				Value:     1.0,
			}},
		}},
		Partition: "1",
	}

	assert.Nil(t, nsqp.push(event))
}

// nsq 0.3.8
func mockNsq() {
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		var count int
		var data []byte

		buf := make([]byte, 16)
		for {
			len, err := r.Body.Read(buf)
			if err != nil && err != io.EOF {
				break
			}
			count += len
			data = append(data, buf[:len]...)
			if err == io.EOF {
				break
			}
		}

		log.Printf("nsq data [%v]\n", string(data))

		w.Write([]byte("ok\n"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
