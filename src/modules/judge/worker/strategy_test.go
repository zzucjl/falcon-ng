package worker

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"
)

// 依赖下游, 不能写单测
func Test_Strategy(t *testing.T) {
	identity = "127.0.0.1"
	opts := NewStrategyConfigOption([]string{"127.0.0.1:8022"})
	ss, err := GetStrategyFromRemote(opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	bytes, _ := json.Marshal(ss)
	fmt.Println(string(bytes))
}

func Test_Yaml(t *testing.T) {
	opts := NewDefaultOptions([]string{"127.0.0.1:1234"},
		[]string{"http://127.0.0.1:8080/api/v1/counter/clude"},
		[]string{"127.0.0.1:8081"}, // url信息单独写在配置文件中
		[]string{"127.0.0.1:6379"})
	bytes, _ := yaml.Marshal(opts)
	fmt.Println(string(bytes))
}
