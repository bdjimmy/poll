package pool

import (
	"net"
	"sync"
)

var _ net.Conn = new(PoolConn)

// PoolConn 实现对net.Conn Close方法的Hook
type PoolConn struct {
	net.Conn
	fd       int
	mu       sync.RWMutex
	c        *channelPool
	unusable bool
}

func (p *PoolConn) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.unusable {
		if p.Conn != nil {
			return p.Conn.Close()
		}
		return nil
	}
	return p.c.put(p)
}

func (p *PoolConn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

func (c *channelPool) wrapConn(conn net.Conn) *PoolConn {
	p := &PoolConn{
		c:  c,
		fd: convertFd(conn),
	}
	p.Conn = conn
	return p
}

func convertFd(conn net.Conn) int {
	f, err := conn.(*net.TCPConn).File()
	if err != nil {
		panic(err)
	}
	return int(f.Fd())
}
