package pool

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	InitialCap = 5
	MaximumCap = 30
	network    = "tcp"
	address    = "127.0.0.1:8989"
	factory    = func() (net.Conn, error) { return net.Dial(network, address) }
)

func TestMain(m *testing.M) {
	go simpleTCPServer()                   // init server
	time.Sleep(time.Millisecond * 300)     // wait until tcp server has been settled
	rand.Seed(time.Now().UTC().UnixNano()) // random
	m.Run()
}

func TestNew(t *testing.T) {
	wg := sync.WaitGroup{}
	pool, err := newChannelPool()
	fmt.Println("init success")

	wg.Add(1)
	go func() {
		fmt.Println(pool.Get())
		select {}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			fmt.Printf("len=%d\n", pool.Len())
			time.Sleep(time.Second)
		}
	}()
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	wg.Wait()
}

func newChannelPool() (Pool, error) {
	return NewChannelPool(InitialCap, MaximumCap, factory)
}

func simpleTCPServer() {
	l, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			time.Sleep(time.Second * 6)
			fmt.Println("start close")
			conn.Close()
		}()
	}
}
