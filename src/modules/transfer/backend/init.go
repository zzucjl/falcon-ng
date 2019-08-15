package backend

import (
	"sync"
	"time"

	"github.com/toolkits/pkg/consistent"
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/container/set"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/pool"
	"github.com/toolkits/pkg/str"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	. "github.com/open-falcon/falcon-ng/src/modules/transfer/config"
)

const (
	DefaultSendQueueMaxSize = 102400 //10.24w
)

var (
	// 服务节点的一致性哈希环 pk -> node
	JudgeHashRing *ConsistentHashRing
	TsdbNodeRing  *ConsistentHashRing

	// 发送缓存队列 node -> queue_of_data
	JudgeQueues = make(map[string]*list.SafeListLimited)
	TsdbQueues  = make(map[string]*list.SafeListLimited)

	// 连接池 node_address -> connection_pool
	JudgeConnPools *ConnPools = &ConnPools{M: make(map[string]*pool.ConnPool)}
	TsdbConnPools  *ConnPools = &ConnPools{M: make(map[string]*pool.ConnPool)}

	JudeNodeHealth     = make(map[string]bool)
	JudeNodeHealthLock = new(sync.RWMutex)

	connTimeout int32
	callTimeout int32
)

func Init() {
	// 初始化默认参数
	connTimeout = int32(Config.Tsdb.ConnTimeout)
	callTimeout = int32(Config.Tsdb.CallTimeout)

	MinStep = Config.MinStep
	if MinStep < 1 {
		MinStep = 10 //默认10s
	}

	initHashRing()
	initConnPools()
	initSendQueues()

	startSendTasks()
	if Config.Judge.Enabled {
		go checkJudgeNodes()
	}
}

func initHashRing() {
	JudgeHashRing = NewConsistentHashRing(int32(Config.Judge.Replicas), str.KeysOfMap(Config.Judge.Cluster))
	TsdbNodeRing = NewConsistentHashRing(int32(Config.Tsdb.Replicas), str.KeysOfMap(Config.Tsdb.Cluster))
}

func initConnPools() {
	var addrs []string
	for _, addr := range Config.Judge.Cluster {
		addrs = append(addrs, addr)
	}
	JudgeConnPools = CreateConnPools(Config.Judge.MaxConns, Config.Judge.MaxIdle,
		Config.Judge.ConnTimeout, Config.Judge.CallTimeout, addrs)

	tsdbInstances := set.NewSafeSet()
	for _, item := range Config.Tsdb.ClusterList {
		for _, addr := range item.Addrs {
			tsdbInstances.Add(addr)
		}
	}
	TsdbConnPools = CreateConnPools(Config.Tsdb.MaxConns, Config.Tsdb.MaxIdle,
		Config.Tsdb.ConnTimeout, Config.Tsdb.CallTimeout, tsdbInstances.ToSlice())
}

func initSendQueues() {
	for node := range Config.Judge.Cluster {
		Q := list.NewSafeListLimited(DefaultSendQueueMaxSize)
		JudgeQueues[node] = Q
	}

	for node, item := range Config.Tsdb.ClusterList {
		for _, addr := range item.Addrs {
			Q := list.NewSafeListLimited(DefaultSendQueueMaxSize)
			TsdbQueues[node+addr] = Q
		}
	}
}

func checkJudgeNodes() {
	if !Config.Judge.Enabled {
		return
	}

	t1 := time.NewTicker(time.Duration(40 * time.Second))
	for {
		<-t1.C
		var wg sync.WaitGroup
		for node, addr := range Config.Judge.Cluster {
			wg.Add(1)
			go checkJudgeNode(node, addr, &wg)
		}
		wg.Wait()
		JudeNodeHealthLock.Lock()
		if len(JudeNodeHealth) > 0 || len(JudgeHashRing.GetRing().Members()) != len(Config.Judge.Cluster) {
			//重建judge hash环
			r := consistent.New()
			r.NumberOfReplicas = Config.Judge.Replicas
			for node, _ := range Config.Judge.Cluster {
				if good, exists := JudeNodeHealth[node]; exists && !good {
					continue
				}
				r.Add(node)
			}
			logger.Warning("judge hash ring rebuild ", r.Members())
			JudgeHashRing.Set(r)
			JudeNodeHealth = make(map[string]bool)
		}
		JudeNodeHealthLock.Unlock()
	}
	return
}

func checkJudgeNode(node, addr string, wg *sync.WaitGroup) {
	defer wg.Done()
	errNum := 0
	resp := &dataobj.SimpleRpcResponse{}
	for i := 0; i < 10; i++ {
		err := JudgeConnPools.Call(addr, "Judge.Ping", "ping", &resp)
		if err != nil {
			errNum += 1
			logger.Warningf("judge[%s] down #%d, err: %v", addr, errNum, err)
			time.Sleep(time.Second)
		}
	}
	if errNum == 10 {
		JudeNodeHealthLock.Lock()
		JudeNodeHealth[node] = false
		JudeNodeHealthLock.Unlock()
		logger.Warning("judge down ", addr)
	}
}
