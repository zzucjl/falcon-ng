package utils

import (
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/nux"

	"github.com/open-falcon/falcon-ng/src/modules/logcollector/config"
)

func CalculateMemLimit() int {
	m, err := nux.MemInfo()
	var memTotal, memLimit int
	if err != nil {
		logger.Error("failed to get mem.total:", err)
		memLimit = 1024
	} else {
		memTotal = int(m.MemTotal / (1024 * 1024))
		memLimit = int(float64(memTotal) * config.Config.MaxMemRate)
	}

	if memLimit < 1024 {
		memLimit = 1024
	}

	return memLimit
}
