package funcs

import (
	"fmt"
	"strings"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/nux"
)

func DeviceMetrics() (L []*dataobj.MetricValue) {
	mountPoints, err := nux.ListMountPoint()
	if err != nil {
		logger.Error("collect device metrics fail:", err)
		return
	}

	var myMountPoints map[string]bool = make(map[string]bool)
	if len(config.Config.Collector.MountPoint) > 0 {
		for _, mp := range config.Config.Collector.MountPoint {
			myMountPoints[mp] = true
		}
	}

	ignoreMountPointsPrefix := config.Config.Collector.MountIgnorePrefix

	var diskTotal uint64 = 0
	var diskUsed uint64 = 0

	for idx := range mountPoints {
		fsSpec, fsFile, fsVfstype := mountPoints[idx][0], mountPoints[idx][1], mountPoints[idx][2]
		if len(myMountPoints) > 0 {
			if _, ok := myMountPoints[fsFile]; !ok {
				logger.Debug("mount point not matched with config", fsFile, "ignored.")
				continue
			}
		}

		if hasIgnorePrefix(fsFile, ignoreMountPointsPrefix) {
			continue
		}

		var du *nux.DeviceUsage
		du, err = nux.BuildDeviceUsage(fsSpec, fsFile, fsVfstype)
		if err != nil {
			logger.Errorf("fsSpec: %s, fsFile: %s, fsVfstype: %s, error: %v", fsSpec, fsFile, fsVfstype, err)
			continue
		}

		if du.BlocksAll == 0 {
			continue
		}

		diskTotal += du.BlocksAll
		diskUsed += du.BlocksUsed

		tags := fmt.Sprintf("mount=%s", du.FsFile)
		L = append(L, GaugeValue("disk.bytes.total", du.BlocksAll, tags))
		L = append(L, GaugeValue("disk.bytes.free", du.BlocksFree, tags))
		L = append(L, GaugeValue("disk.bytes.used", du.BlocksUsed, tags))
		L = append(L, GaugeValue("disk.bytes.used.percent", du.BlocksUsedPercent, tags))

		if du.InodesAll == 0 {
			continue
		}

		L = append(L, GaugeValue("disk.inodes.total", du.InodesAll, tags))
		L = append(L, GaugeValue("disk.inodes.free", du.InodesFree, tags))
		L = append(L, GaugeValue("disk.inodes.used", du.InodesUsed, tags))
		L = append(L, GaugeValue("disk.inodes.used.percent", du.InodesUsedPercent, tags))
	}

	if len(L) > 0 && diskTotal > 0 {
		L = append(L, GaugeValue("disk.cap.bytes.total", float64(diskTotal)))
		L = append(L, GaugeValue("disk.cap.bytes.used", float64(diskUsed)))
		L = append(L, GaugeValue("disk.cap.bytes.free", float64(diskTotal-diskUsed)))
		L = append(L, GaugeValue("disk.cap.bytes.used.percent", float64(diskUsed)*100.0/float64(diskTotal)))
	}

	return
}

func hasIgnorePrefix(fsFile string, ignoreMountPointsPrefix []string) bool {
	hasPrefix := false
	if len(ignoreMountPointsPrefix) > 0 {
		for _, ignorePrefix := range ignoreMountPointsPrefix {
			if strings.HasPrefix(fsFile, ignorePrefix) {
				hasPrefix = true
				logger.Debugf("mount point %s has ignored prefix %s", fsFile, ignorePrefix)
				break
			}
		}
	}
	return hasPrefix
}
