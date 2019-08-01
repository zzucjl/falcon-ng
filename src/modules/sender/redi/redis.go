package redi

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/open-falcon/falcon-ng/src/modules/sender/config"
)

var RedisConnPool *redis.Pool

var (
	SMS_QUEUE_NAME   string
	IM_QUEUE_NAME    string
	VOICE_QUEUE_NAME string
	MAIL_QUEUE_NAME  string
)

func InitRedis() {
	cfg := config.GetCfgYml()

	addr := cfg.Redis.Addr
	pass := cfg.Redis.Pass
	maxIdle := cfg.Redis.Idle
	idleTimeout := 240 * time.Second

	connTimeout := time.Duration(cfg.Redis.Timeout.Conn) * time.Millisecond
	readTimeout := time.Duration(cfg.Redis.Timeout.Read) * time.Millisecond
	writeTimeout := time.Duration(cfg.Redis.Timeout.Write) * time.Millisecond

	SMS_QUEUE_NAME = cfg.Queue.Sms
	IM_QUEUE_NAME = cfg.Queue.IM
	MAIL_QUEUE_NAME = cfg.Queue.Mail
	VOICE_QUEUE_NAME = cfg.Queue.Voice

	RedisConnPool = &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr, redis.DialConnectTimeout(connTimeout), redis.DialReadTimeout(readTimeout), redis.DialWriteTimeout(writeTimeout))
			if err != nil {
				return nil, err
			}

			if pass != "" {
				if _, err := c.Do("AUTH", pass); err != nil {
					c.Close()
					log.Println("ERR: redis auth fail")
					return nil, err
				}
			}

			return c, err
		},
		TestOnBorrow: PingRedis,
	}
}

func PingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	if err != nil {
		log.Println("ERR: ping redis fail", err)
	}
	return err
}

func CloseRedis() {
	log.Println("INFO: closing redis...")
	RedisConnPool.Close()
}

func ChoiceQueue(notifyType string) string {
	switch notifyType {
	case "im":
		return IM_QUEUE_NAME
	case "mail":
		return MAIL_QUEUE_NAME
	case "sms":
		return SMS_QUEUE_NAME
	case "voice":
		return VOICE_QUEUE_NAME
	}

	return ""
}
