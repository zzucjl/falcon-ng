package redis

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"

	"github.com/gomodule/redigo/redis"
)

type RedisPublisher struct {
	opts   publish.RedisPublisherOption
	conns  []*redis.Pool
	buffer chan *schema.Event
	stop   chan struct{}
	closed bool
}

const (
	Balance_Round_Robbin = "round_robbin" // 轮询
	Balance_Random       = "random"       // 随机
)

func NewRedisPublisher(opts publish.RedisPublisherOption) (*RedisPublisher, error) {
	if len(opts.Addrs) == 0 {
		return nil, errors.New("empty redis addr")
	}
	if len(opts.Balance) == 0 {
		opts.Balance = Balance_Random
	}
	if opts.Balance != Balance_Round_Robbin &&
		opts.Balance != Balance_Random {
		return nil, errors.New("unsupported redis load balance")
	}

	if opts.BufferSize == 0 {
		opts.BufferSize = 1
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = 240000 // 240s
	}

	redisp := &RedisPublisher{
		opts:   opts,
		conns:  make([]*redis.Pool, 0),
		buffer: make(chan *schema.Event, opts.BufferSize),
		stop:   make(chan struct{}, 1),
		closed: false,
	}
	for i := range opts.Addrs {
		rpool := &redis.Pool{
			MaxIdle:     opts.MaxIdle,
			IdleTimeout: publish.Duration(opts.IdleTimeout),
			Dial: func() (redis.Conn, error) {
				c, err := redis.DialTimeout("tcp",
					opts.Addrs[i],
					publish.Duration(opts.ConnTimeout),
					publish.Duration(opts.ReadTimeout),
					publish.Duration(opts.WriteTimeout))

				if err != nil {
					return nil, err
				}
				if len(opts.Password) > 0 {
					if _, err := c.Do("AUTH", opts.Password); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("ping")
				if err != nil {
					logger.Errorf(0, "ping redis failed: %v", err)
				}
				return err
			},
		}
		conn := rpool.Get()
		_, err := conn.Do("ping")
		conn.Close()
		if err == nil {
			redisp.conns = append(redisp.conns, rpool)
		}
	}
	if len(redisp.conns) == 0 {
		return nil, errors.New("redis server not available")
	}

	go redisp.loop()

	return redisp, nil
}

func (rp *RedisPublisher) Publish(event *schema.Event) error {
	if rp.closed {
		return nil
	}

	var err error
	select {
	case rp.buffer <- event:
		// do nothing
	case <-time.After(publish.Duration(rp.opts.BufferEnqueueTimeout)):
		// 入队列超时, 直接写入到redis
		err = rp.push(event)
	}

	return err
}

func (rp *RedisPublisher) Close() {
	rp.closed = true
	rp.stop <- struct{}{}

	for {
		done := false
		select {
		case event := <-rp.buffer:
			rp.push(event)
		case <-time.After(time.Millisecond * 50):
			done = true
		}
		if done {
			break
		}
	}

	logger.Info(0, "redis publish closed")
	return
}

func (rp *RedisPublisher) loop() {
	for {
		select {
		case <-rp.stop:
			logger.Info(0, "redis publish loop stopped")
			return
		case event := <-rp.buffer:
			rp.push(event)
		}
	}
}

func (rp *RedisPublisher) push(event *schema.Event) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		logger.Warningf(0, "redis publish failed, error:%v", err)
		return err
	}

	succ := false
	pools := rp.selectRedisPools()
	if len(pools) == 0 {
		logger.Warningf(event.Sid, "redis publish failed: empty conn pools")
		return errors.New("redis publish failed")
	}
	for i := range pools {
		rc := pools[i].Get()
		defer rc.Close()

		// 写入用lpush 读出应该用 rpop
		_, err = rc.Do("RPUSH", event.Partition, string(bytes))

		if err == nil {
			succ = true
			break
		}
		logger.Debugf(event.Sid, "redis publish failed, error:%v", err)
	}

	if succ {
		logger.Debugf(event.Sid, "redis publish succ, event: %s", string(bytes))
		return nil
	}

	logger.Warningf(event.Sid, "redis publish failed finally")
	return errors.New("redis publish failed")
}

func (rp *RedisPublisher) selectRedisPools() []*redis.Pool {
	var ret []*redis.Pool
	if len(rp.conns) == 0 {
		return ret
	}
	switch rp.opts.Balance {
	case Balance_Random:
		rand.Seed(time.Now().Unix())
		perm := rand.Perm(len(rp.conns))
		for i := range perm {
			ret = append(ret, rp.conns[i])
		}
	case Balance_Round_Robbin:
		for i := range rp.conns {
			ret = append(ret, rp.conns[i])
		}
	}

	return ret
}
