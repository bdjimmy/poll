package pool

import (
	"errors"
	"net"
)

var (
	// ErrorClosed is the error resulting if the pool is closed via pool.Close()
	ErrClosed = errors.New("pool is close")
)

// Factory is a function to create new connections
type Factory func() (net.Conn, error)

// Pool interface describes a pool inplementation.
// A pool should have maximum capacity. An ideal pool is threadsafe and easy to use.
type Pool interface {
	// Get return a new connection from the pool.
	Get() (net.Conn, error)
	// Close closes the pool and all its connections.
	// After Close() the pool is no longer usable.
	Close()
	// LEn returns the current number of connections of pool.
	Len() int
}
