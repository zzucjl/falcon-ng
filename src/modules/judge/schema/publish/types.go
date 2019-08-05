package publish

import (
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
)

// EventPublisher
type EventPublisher interface {
	Publish(event *schema.Event) error
	Close() // 如果publisher带有缓冲区, 执行Close() 代表 关闭写入, 尽可能写出
}
