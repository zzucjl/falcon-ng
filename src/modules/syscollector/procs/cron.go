package procs

import (
	"time"
)

func Detect() {
	detect()
	go loopDetect()
}

func loopDetect() {
	for {
		time.Sleep(time.Second * 10)
		detect()
	}
}

func detect() {
	ps := ListProcs()
	DelNoPorcCollect(ps)
	AddNewPorcCollect(ps)
}
