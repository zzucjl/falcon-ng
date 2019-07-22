package nsq

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"

	"github.com/parnurzeal/gorequest"
)

// NsqPublisher
type NsqPublisher struct {
	opts   publish.NsqPublisherOption
	client *gorequest.SuperAgent
	buffer chan *schema.Event
	stop   chan struct{}
	closed bool
}

func NewNsqPublisher(opts publish.NsqPublisherOption) (*NsqPublisher, error) {
	if len(opts.Addrs) == 0 {
		return nil, errors.New("empty nsq addr")
	}
	client := gorequest.New().Timeout(publish.Duration(opts.CallTimeout))
	if client == nil {
		return nil, errors.New("nsq client init failed")
	}
	if opts.BufferSize == 0 {
		opts.BufferSize = 1
	}

	nsqp := &NsqPublisher{
		opts:   opts,
		client: client,
		buffer: make(chan *schema.Event, opts.BufferSize),
		stop:   make(chan struct{}, 1),
		closed: false,
	}
	go nsqp.loop()

	return nsqp, nil
}

func (np *NsqPublisher) Publish(event *schema.Event) error {
	if np.closed {
		return nil
	}

	var err error
	select {
	case np.buffer <- event:
		// do nothing
	case <-time.After(publish.Duration(np.opts.BufferEnqueueTimeout)):
		// 入队列超时, 直接报错
		// gorequest 的 client 并发使用会panic(map concurrency read and write问题)
		err = errors.New("buffer is full")
	}

	return err
}

func (np *NsqPublisher) Close() {
	np.closed = true
	np.stop <- struct{}{}

	for event := range np.buffer {
		np.push(event)
	}
	close(np.buffer)
	logger.Info(0, "nsq publish closed")
	return
}

func (np *NsqPublisher) loop() {
	for {
		select {
		case <-np.stop:
			logger.Info(0, "nsq publish loop stopped")
			return
		case event := <-np.buffer:
			np.push(event)
		}
	}
}

func (np *NsqPublisher) push(event *schema.Event) error {
	succ := false
	perm := rand.Perm(len(np.opts.Addrs))
	for i := range perm {
		resp, body, err := np.client.
			Post(np.opts.Addrs[perm[i]] + "?topic=" + event.Partition).
			Type("json").
			Send(event).
			End()

		var code int
		if len(err) != 0 {
			logger.Debugf(event.Sid, "nsq publish failed, error:%v", err)
			continue
		}

		if resp != nil {
			code = resp.StatusCode
		}

		if code == http.StatusOK {
			succ = true
			break
		}

		logger.Debugf(event.Sid, "nsq publish failed, code:%d, body:%s, error:%v",
			code, body, err)
	}

	if succ {
		return nil
	}

	logger.Warningf(event.Sid, "nsq publish failed finally")
	return errors.New("nsq publish failed")
}
