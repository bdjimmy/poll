package pool

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"
)

type channelPool struct {
	mu      sync.RWMutex
	factory Factory
	maxCap  int
	epoll   *epollM
	maps    map[int]*list.Element
	list    *list.List
}

func NewChannelPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}
	poll, err := createEpoll()
	if err != nil {
		return nil, errors.New(err.Error())
	}
	pool := &channelPool{
		factory: factory,
		maxCap:  maxCap,
		epoll:   poll,
		maps:    make(map[int]*list.Element, maxCap),
		list:    list.New(),
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := pool.create()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		conn.Close()
	}

	// epool wait
	go func() {
		poll.wait(func(fd int32) {
			pool.mu.Lock()
			defer pool.mu.Unlock()

			if e, ok := pool.maps[int(fd)]; ok {
				pool.list.Remove(e)
				delete(pool.maps, int(fd))
			}

			conn, err := pool.create()
			if err == nil {
				conn.Close()
			}

		})
	}()

	return pool, nil
}

func (c *channelPool) Get() (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.list == nil {
		return nil, ErrClosed
	}

	element := c.list.Front()
	if element != nil {
		// delete epoll
		c.epoll.removeEvent(element.Value.(*PoolConn))
		c.list.Remove(element)
		delete(c.maps, element.Value.(*PoolConn).fd)
		return element.Value.(*PoolConn), nil
	}

	return c.create()
}

func (c *channelPool) create() (net.Conn, error) {
	// create a connection
	conn, err := c.factory()
	if err != nil {
		return nil, err
	}
	wConn := c.wrapConn(conn)
	return wConn, nil
}

func (c *channelPool) put(conn *PoolConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	if c.list == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.

	if c.list.Len() >= c.maxCap {
		return conn.Close()
	}

	// add event
	if err := c.epoll.addEvent(conn); err != nil {
		return fmt.Errorf("add event err:%s", err)
	}
	c.maps[conn.fd] = c.list.PushBack(conn)
	return nil
}

func (c *channelPool) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	maps := c.maps
	c.maps = nil
	c.list = nil
	c.epoll = nil
	c.factory = nil
	c.mu.Unlock()
	if maps == nil {
		return
	}
	for fd := range maps {
		syscall.Close(fd)
	}
}

func (c *channelPool) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.maps)
}
