package query

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"

	codec "github.com/ugorji/go/codec"
)

// RpcCient, 要实现io.Closer接口
type RpcClient struct {
	cli  *rpc.Client
	name string
}

func (c RpcClient) Name() string {
	return c.name
}

func (c RpcClient) Closed() bool {
	return c.cli == nil
}

func (c RpcClient) Close() error {
	if c.cli != nil {
		err := c.cli.Close()
		c.cli = nil
		return err
	}
	return nil
}

func (c RpcClient) Call(method string, args interface{}, reply interface{}) error {
	return c.cli.Call(method, args, reply)
}

// ConnPools Manager
type SafeRpcConnPools struct {
	sync.RWMutex
	M           map[string]*ConnPool
	MaxConns    int
	MaxIdle     int
	ConnTimeout int
	CallTimeout int
}

func CreateSafeRpcConnPools(maxConns, maxIdle, connTimeout, callTimeout int, cluster []string) *SafeRpcConnPools {
	cp := &SafeRpcConnPools{M: make(map[string]*ConnPool), MaxConns: maxConns, MaxIdle: maxIdle,
		ConnTimeout: connTimeout, CallTimeout: callTimeout}

	ct := time.Duration(cp.ConnTimeout) * time.Millisecond
	for _, address := range cluster {
		if _, exist := cp.M[address]; exist {
			continue
		}
		cp.M[address] = createOnePool(address, address, ct, maxConns, maxIdle)
	}

	return cp
}

func CreateSafeRpcWithCodecConnPools(maxConns, maxIdle, connTimeout, callTimeout int, cluster []string) *SafeRpcConnPools {
	cp := &SafeRpcConnPools{M: make(map[string]*ConnPool), MaxConns: maxConns, MaxIdle: maxIdle,
		ConnTimeout: connTimeout, CallTimeout: callTimeout}

	ct := time.Duration(cp.ConnTimeout) * time.Millisecond
	for _, address := range cluster {
		if _, exist := cp.M[address]; exist {
			continue
		}
		cp.M[address] = createOneWithCodecPool(address, address, ct, maxConns, maxIdle)
	}

	return cp
}

func (spool *SafeRpcConnPools) Proc() []string {
	procs := []string{}
	for _, cp := range spool.M {
		procs = append(procs, cp.Proc())
	}
	return procs
}

// 同步发送, 完成发送或超时后 才能返回
func (spool *SafeRpcConnPools) Call(addr, method string, args interface{}, resp interface{}) error {
	connPool, exists := spool.Get(addr)
	if !exists {
		return fmt.Errorf("%s has no connection pool", addr)
	}

	conn, err := connPool.Fetch()
	if err != nil {
		return fmt.Errorf("%s get connection fail: conn %v, err %v. proc: %s", addr, conn, err, connPool.Proc())
	}

	rpcClient := conn.(RpcClient)
	callTimeout := time.Duration(spool.CallTimeout) * time.Millisecond

	done := make(chan error, 1)
	go func() {
		done <- rpcClient.Call(method, args, resp)
	}()

	select {
	case <-time.After(callTimeout):
		connPool.ForceClose(conn)
		return fmt.Errorf("%s, call timeout", addr)
	case err = <-done:
		if err != nil {
			connPool.ForceClose(conn)
			err = fmt.Errorf("%s, call failed, err %v. proc: %s", addr, err, connPool.Proc())
		} else {
			connPool.Release(conn)
		}
		return err
	}
}

func (spool *SafeRpcConnPools) Get(address string) (*ConnPool, bool) {
	spool.RLock()
	defer spool.RUnlock()
	p, exists := spool.M[address]
	return p, exists
}

func (spool *SafeRpcConnPools) Destroy() {
	spool.Lock()
	defer spool.Unlock()
	addresses := make([]string, 0, len(spool.M))
	for address := range spool.M {
		addresses = append(addresses, address)
	}

	for _, address := range addresses {
		spool.M[address].Destroy()
		delete(spool.M, address)
	}
}

func createOnePool(name string, address string, connTimeout time.Duration, maxConns int, maxIdle int) *ConnPool {
	p := NewConnPool(name, address, maxConns, maxIdle)
	p.New = func(connName string) (NConn, error) {
		_, err := net.ResolveTCPAddr("tcp", p.Address)
		if err != nil {
			//log.Println(p.Address, "format error", err)
			return nil, err
		}

		conn, err := net.DialTimeout("tcp", p.Address, connTimeout)
		if err != nil {
			//log.Printf("new conn fail, addr %s, err %v", p.Address, err)
			return nil, err
		}

		return RpcClient{cli: rpc.NewClient(conn), name: connName}, nil
	}

	return p
}

func createOneWithCodecPool(name string, address string, connTimeout time.Duration, maxConns int, maxIdle int) *ConnPool {
	p := NewConnPool(name, address, maxConns, maxIdle)
	p.New = func(connName string) (NConn, error) {
		_, err := net.ResolveTCPAddr("tcp", p.Address)
		if err != nil {
			//log.Println(p.Address, "format error", err)
			return nil, err
		}

		conn, err := net.DialTimeout("tcp", p.Address, connTimeout)
		if err != nil {
			//log.Printf("new conn fail, addr %s, err %v", p.Address, err)
			return nil, err
		}

		var h codec.MsgpackHandle
		rpcCodec := codec.MsgpackSpecRpc.ClientCodec(conn, &h)
		return RpcClient{cli: rpc.NewClientWithCodec(rpcCodec), name: connName}, nil
	}

	return p
}

// TSDB
type TsdbClient struct {
	cli  net.Conn
	name string
}

func (c TsdbClient) Name() string {
	return c.name
}

func (c TsdbClient) Closed() bool {
	return c.cli == nil
}

func (c TsdbClient) Close() error {
	if c.cli != nil {
		err := c.cli.Close()
		c.cli = nil
		return err
	}
	return nil
}

func newTsdbConnPool(address string, maxConns int, maxIdle int, connTimeout int) *ConnPool {
	pool := NewConnPool("tsdb", address, maxConns, maxIdle)

	pool.New = func(name string) (NConn, error) {
		_, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			return nil, err
		}

		conn, err := net.DialTimeout("tcp", address, time.Duration(connTimeout)*time.Millisecond)
		if err != nil {
			return nil, err
		}

		return TsdbClient{conn, name}, nil
	}

	return pool
}

type TsdbConnPoolHelper struct {
	p           *ConnPool
	maxConns    int
	maxIdle     int
	connTimeout int
	callTimeout int
	address     string
}

func NewTsdbConnPoolHelper(address string, maxConns, maxIdle, connTimeout, callTimeout int) *TsdbConnPoolHelper {
	return &TsdbConnPoolHelper{
		p:           newTsdbConnPool(address, maxConns, maxIdle, connTimeout),
		maxConns:    maxConns,
		maxIdle:     maxIdle,
		connTimeout: connTimeout,
		callTimeout: callTimeout,
		address:     address,
	}
}

func (helper *TsdbConnPoolHelper) Send(data []byte) (err error) {
	conn, err := helper.p.Fetch()
	if err != nil {
		return fmt.Errorf("get connection fail: err %v. proc: %s", err, helper.p.Proc())
	}

	cli := conn.(TsdbClient).cli

	done := make(chan error)
	go func() {
		_, err = cli.Write(data)
		done <- err
	}()

	select {
	case <-time.After(time.Duration(helper.callTimeout) * time.Millisecond):
		helper.p.ForceClose(conn)
		return fmt.Errorf("%s, call timeout", helper.address)
	case err = <-done:
		if err != nil {
			helper.p.ForceClose(conn)
			err = fmt.Errorf("%s, call failed, err %v. proc: %s", helper.address, err, helper.p.Proc())
		} else {
			helper.p.Release(conn)
		}
		return err
	}

	return
}

func (helper *TsdbConnPoolHelper) Destroy() {
	if helper.p != nil {
		helper.p.Destroy()
	}
}
