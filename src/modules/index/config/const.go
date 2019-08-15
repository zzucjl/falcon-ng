package config

import "errors"

const Version = 1

var (
	ErrNidNotFound          = errors.New("cannot find this nid")
	ErrNidMetricNotFound    = errors.New("cannot find this metric in target nid")
	ErrNidMetricTagNotFound = errors.New("cannot find this tagkey in target nid & metric")
	ErrDupTagName           = errors.New("the tagName must be unique")
	ErrTooManyTags          = errors.New("too many tags")
	ErrEmptyTagk            = errors.New("empty tagk")
)

var (
	DEFAULT_METRIC = map[string]struct{}{
		"proc.agent.alive":                             struct{}{},
		"proc.agent.version":                           struct{}{},
		"cpu.guest":                                    struct{}{},
		"cpu.idle":                                     struct{}{},
		"cpu.iowait":                                   struct{}{},
		"cpu.irq":                                      struct{}{},
		"cpu.loadavg.1":                                struct{}{},
		"cpu.loadavg.15":                               struct{}{},
		"cpu.loadavg.5":                                struct{}{},
		"cpu.nice":                                     struct{}{},
		"cpu.softirq":                                  struct{}{},
		"cpu.steal":                                    struct{}{},
		"cpu.switches":                                 struct{}{},
		"cpu.sys":                                      struct{}{},
		"cpu.user":                                     struct{}{},
		"cpu.util":                                     struct{}{},
		"disk.cap.bytes.total":                         struct{}{},
		"disk.cap.bytes.used":                          struct{}{},
		"disk.cap.bytes.used.percent":                  struct{}{},
		"mem.bytes.buffers":                            struct{}{},
		"mem.bytes.cached":                             struct{}{},
		"mem.bytes.free":                               struct{}{},
		"mem.bytes.total":                              struct{}{},
		"mem.bytes.used":                               struct{}{},
		"mem.bytes.used.percent":                       struct{}{},
		"net.bandwidth.mbits.total":                    struct{}{},
		"net.in.bits.total":                            struct{}{},
		"net.in.bits.total.percent":                    struct{}{},
		"net.out.bits.total":                           struct{}{},
		"net.out.bits.total.percent":                   struct{}{},
		"net.sockets.tcp.inuse":                        struct{}{},
		"net.sockets.tcp.timewait":                     struct{}{},
		"net.sockets.used":                             struct{}{},
		"sys.fs.files.free":                            struct{}{},
		"sys.fs.files.max":                             struct{}{},
		"sys.fs.files.used":                            struct{}{},
		"sys.fs.files.used.percent":                    struct{}{},
		"sys.net.netfilter.nf_conntrack_count":         struct{}{},
		"sys.net.netfilter.nf_conntrack_count.percent": struct{}{},
		"sys.net.netfilter.nf_conntrack_max":           struct{}{},
		"sys.ntp.offset.ms":                            struct{}{},
		"sys.ps.entity.total":                          struct{}{},
		"sys.ps.process.total":                         struct{}{},
	}

	DEFAULT_DSTYPE = "GAUGE"
	DEFAULT_STEP   = 10
)
