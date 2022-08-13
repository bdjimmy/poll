//go:build linux

package pool

import (
	"syscall"
)

type epollM struct {
	epollFd int
}

func createEpoll() (*epollM, error) {
	fd, err := syscall.EpollCreate(1)
	if err != nil {
		return nil, err
	}
	return &epollM{epollFd: fd}, nil
}

func (e *epollM) addEvent(conn *PoolConn) error {
	// syscall.EPOLLHUP syscall.EPOLLERR
	// Error condition happened on the associated file descriptor.
	// epoll_wait(2) will always wait for this event; it is not
	// necessary to set it in events.

	// EPOLLRDHUP (since Linux 2.6.17)
	// Stream socket peer closed connection, or shut down writing half of connection.
	// This flag  is  especially  useful for writing simple code to detect peer shutdown
	// when using Edge Triggered monitoring.

	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, conn.fd, &syscall.EpollEvent{
		Events: syscall.EPOLLRDHUP | -syscall.EPOLLET,
		Fd:     int32(conn.fd),
		Pad:    0,
	})
}

func (e *epollM) removeEvent(conn *PoolConn) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, conn.fd, nil)
}

func (e *epollM) close() error {
	return syscall.Close(e.epollFd)
}

func (e *epollM) wait(handler func(fd int32)) error {
	events := make([]syscall.EpollEvent, 100)
	for {
		n, err := syscall.EpollWait(e.epollFd, events, -1)
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			if events[i].Events&syscall.EPOLLHUP == syscall.EPOLLHUP ||
				events[i].Events&syscall.EPOLLERR == syscall.EPOLLERR ||
				events[i].Events&syscall.EPOLLRDHUP == syscall.EPOLLRDHUP {
				handler(events[i].Fd)
			}
		}
	}
}
