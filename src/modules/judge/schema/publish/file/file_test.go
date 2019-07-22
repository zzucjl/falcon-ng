package file

import (
	"testing"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/stretchr/testify/assert"
)

func Test_Log(t *testing.T) {
	logp, _ := NewFilePublisher(publish.FilePublisherOption{
		Name: "./log",
	})
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
	assert.Nil(t, logp.Publish(event))

	logp.Close()
}
