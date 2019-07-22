package series

import (
	"bytes"
	"sort"
	"strings"
	"sync"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} { return new(bytes.Buffer) },
	}
)

// Map2SortedString map转化为字符串, 不同key之间用逗号分隔, key按字母序排列
// key = "endpoint" 直接跳过
func Map2SortedString(tags map[string]string) string {
	if tags == nil || len(tags) == 0 {
		return ""
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	keys := make([]string, 0)
	for key := range tags {
		if key == schema.ENDPOINT_KEYWORD {
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for i := range keys {
		buf.WriteString(keys[i])
		buf.WriteString("=")
		buf.WriteString(tags[keys[i]])
		if i != len(keys)-1 {
			buf.WriteString(",")
		}
	}

	return buf.String()
}

// CounterString2TagMap index返回的counter字符串转化为map
func CounterString2TagMap(counter string) map[string]string {
	if counter == "" {
		return map[string]string{}
	}

	if strings.ContainsRune(counter, ' ') {
		counter = strings.Replace(counter, " ", "", -1)
	}

	tagM := make(map[string]string)
	for _, tag := range strings.Split(counter, ",") {
		idx := strings.IndexRune(tag, '=')
		if idx != -1 {
			tagM[tag[:idx]] = tag[idx+1:]
		}
	}

	return tagM
}
