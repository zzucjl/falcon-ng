package publish

import (
	"time"
)

var (
	timeoutUnit = time.Millisecond * 1
)

func Duration(n int) time.Duration {
	return time.Duration(n) * timeoutUnit
}

type PublisherOption struct {
	Type  string               `yaml:"type"`
	Nsq   NsqPublisherOption   `yaml:"nsq,omitempty"`
	File  FilePublisherOption  `yaml:"file,omitempty"`
	Redis RedisPublisherOption `yaml:"redis,omitempty"`
}

type NsqPublisherOption struct {
	Addrs                []string `yaml:"addrs"`                // 直连下游nsq的地址
	CallTimeout          int      `yaml:"callTimeout"`          // 请求超时
	BufferSize           int      `yaml:"bufferSize"`           // 缓存个数
	BufferEnqueueTimeout int      `yaml:"bufferEnqueueTimeout"` // 缓存入队列超时
}

type FilePublisherOption struct {
	Name string `yaml:"name"`
}

type RedisPublisherOption struct {
	Addrs                []string `yaml:"addrs"`                // 直连下游redis的地址
	Password             string   `yaml:"password"`             // 密码
	Balance              string   `yaml:"balance"`              // load balance, 负载均衡算法
	ConnTimeout          int      `yaml:"connTimeout"`          // 连接超时
	ReadTimeout          int      `yaml:"readTimeout"`          // 读超时
	WriteTimeout         int      `yaml:"writeTimeout"`         // 写超时
	MaxIdle              int      `yaml:"maxIdle"`              // idle
	IdleTimeout          int      `yaml:"idleTimeout"`          // 超时
	BufferSize           int      `yaml:"bufferSize"`           // 缓存个数
	BufferEnqueueTimeout int      `yaml:"bufferEnqueueTimeout"` // 缓存入队列超时
}

func NewRedisPublisherOption(addrs []string) PublisherOption {
	return PublisherOption{
		Type: "redis",
		Redis: RedisPublisherOption{
			Addrs:                addrs,
			Password:             "",
			ConnTimeout:          200,
			ReadTimeout:          500,
			WriteTimeout:         500,
			MaxIdle:              10,
			IdleTimeout:          100,
			BufferSize:           1024,
			BufferEnqueueTimeout: 200,
		},
	}
}
