package model

import (
	"log"

	"github.com/toolkits/pkg/logger"
)

type Node struct {
	Id   int64  `json:"id"`
	Pid  int64  `json:"pid"`
	Name string `json:"name"`
	Path string `json:"path"`
	Leaf int    `json:"leaf"`
	Type int    `json:"type"`
	Note string `json:"note"`
}

// InitNode 初始化第一个node节点
func InitNode() {
	num, err := DB["portal"].Where("pid=0").Count(new(Node))
	if err != nil {
		log.Fatalln("cannot query portal.node", err)
	}

	if num > 0 {
		return
	}

	node := Node{
		Pid:  0,
		Name: "cop",
		Path: "cop",
		Leaf: 0,
		Type: 0,
		Note: "公司节点",
	}

	_, err = DB["portal"].Insert(&node)
	if err != nil {
		log.Fatalln("cannot insert node[cop]")
	}

	logger.Info("node cop init done")
}
