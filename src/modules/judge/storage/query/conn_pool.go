package query

import (
	"fmt"
	"io"
	"sync"
	"time"
)

var ErrMaxConn = fmt.Errorf("maximum connections reached")

//
type NConn interface {
	io.Closer
	Name() string
	Closed() bool
}

type ConnPool struct {
	sync.RWMutex

	Name     string
	Address  string
	MaxConns int
	MaxIdle  int
	Cnt      int64
	New      func(name string) (NConn, error)

	active int
	free   []NConn
	all    map[string]NConn
}

func NewConnPool(name string, address string, maxConns int, maxIdle int) *ConnPool {
	return &ConnPool{Name: name, Address: address, MaxConns: maxConns, MaxIdle: maxIdle, Cnt: 0, all: make(map[string]NConn)}
}

func (pool *ConnPool) Proc() string {
	pool.RLock()
	defer pool.RUnlock()

	return fmt.Sprintf("Name:%s,Cnt:%d,active:%d,all:%d,free:%d",
		pool.Name, pool.Cnt, pool.active, len(pool.all), len(pool.free))
}

func (pool *ConnPool) Fetch() (NConn, error) {
	pool.Lock()
	defer pool.Unlock()

	// get from free
	conn := pool.fetchFree()
	if conn != nil {
		return conn, nil
	}

	if pool.overMax() {
		return nil, ErrMaxConn
	}

	// create new conn
	conn, err := pool.newConn()
	if err != nil {
		return nil, err
	}

	pool.increActive()
	return conn, nil
}

func (pool *ConnPool) Release(conn NConn) {
	pool.Lock()
	defer pool.Unlock()

	if pool.overMaxIdle() {
		pool.deleteConn(conn)
		pool.decreActive()
	} else {
		pool.addFree(conn)
	}
}

func (pool *ConnPool) ForceClose(conn NConn) {
	pool.Lock()
	defer pool.Unlock()

	pool.deleteConn(conn)
	pool.decreActive()
}

func (pool *ConnPool) Destroy() {
	pool.Lock()
	defer pool.Unlock()

	for _, conn := range pool.free {
		if conn != nil && !conn.Closed() {
			conn.Close()
		}
	}

	for _, conn := range pool.all {
		if conn != nil && !conn.Closed() {
			conn.Close()
		}
	}

	pool.active = 0
	pool.free = []NConn{}
	pool.all = map[string]NConn{}
}

// internal, concurrently unsafe
func (pool *ConnPool) newConn() (NConn, error) {
	name := fmt.Sprintf("%s_%d_%d", pool.Name, pool.Cnt, time.Now().Unix())
	conn, err := pool.New(name)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	pool.Cnt++
	pool.all[conn.Name()] = conn
	return conn, nil
}

func (pool *ConnPool) deleteConn(conn NConn) {
	if conn != nil {
		conn.Close()
	}
	delete(pool.all, conn.Name())
}

func (pool *ConnPool) addFree(conn NConn) {
	pool.free = append(pool.free, conn)
}

func (pool *ConnPool) fetchFree() NConn {
	if len(pool.free) == 0 {
		return nil
	}

	conn := pool.free[0]
	pool.free = pool.free[1:]
	return conn
}

func (pool *ConnPool) increActive() {
	pool.active += 1
}

func (pool *ConnPool) decreActive() {
	pool.active -= 1
}

func (pool *ConnPool) overMax() bool {
	return pool.active >= pool.MaxConns
}

func (pool *ConnPool) overMaxIdle() bool {
	return len(pool.free) >= pool.MaxIdle
}
