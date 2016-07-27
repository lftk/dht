package main

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/4396/dht"
)

func newRandomID() *dht.ID {
	id := new(dht.ID)
	n, err := rand.Read(id[:])
	if err != nil || n != dht.IDLen {
		return dht.ZeroID
	}
	return id
}

type dhtQueryTracker struct {
}

func (t *dhtQueryTracker) Ping(id *dht.ID) {
	fmt.Println("q", "ping", id)
}

func (t *dhtQueryTracker) FindNode(id *dht.ID, target *dht.ID) {
	fmt.Println("q", "find_node", id)
}

func (t *dhtQueryTracker) GetPeers(id *dht.ID, tor *dht.ID) {
	fmt.Println("q", "get_peers", id)
}

func (t *dhtQueryTracker) AnnouncePeer(id *dht.ID, tor *dht.ID, peer []byte) {
	fmt.Println("q", "announce_peer", id)
}

type dhtReplyTracker struct {
}

func (t *dhtReplyTracker) Ping(id *dht.ID) {
	fmt.Println("r", "ping", id)
}

func (t *dhtReplyTracker) FindNode(id *dht.ID, nodes []byte) {
	fmt.Println("r", "find_node", id)
}

func (t *dhtReplyTracker) GetPeers(id *dht.ID, peers [][]byte, nodes []byte) {
	fmt.Println("r", "get_peers", id)
}

func (t *dhtReplyTracker) AnnouncePeer(id *dht.ID) {
	fmt.Println("r", "announce_peer", id)
}

type dhtErrorTracker struct {
}

func (t *dhtErrorTracker) Error(val int, err string) {
	fmt.Println("e", val, err)
}

var routers = []string{
	"router.magnets.im:6881",
	"router.bittorrent.com:6881",
	"dht.transmissionbt.com:6881",
	"router.utorrent.com:6881",
}

func initDHTServer(d *dht.DHT) (err error) {
	for _, addr := range routers {
		addr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			break
		}
		err = d.FindNodeFromAddr(d.ID(), addr)
		if err != nil {
			break
		}
	}
	return
}

type udpMessage struct {
	addr *net.UDPAddr
	data []byte
	size int
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return
	}

	d := dht.NewDHT(newRandomID(), conn.(*net.UDPConn), 16)
	t := dht.NewTracker(&dhtQueryTracker{}, &dhtReplyTracker{}, &dhtErrorTracker{})
	exit := make(chan interface{})
	msg := make(chan *udpMessage, 1024)
	datas := &sync.Pool{New: func() interface{} {
		return make([]byte, 1024)
	}}

	go func(msg chan *udpMessage) {
		if err = initDHTServer(d); err != nil {
			fmt.Println(err)
			close(exit)
			return
		}
		conn := d.Conn()
		buf := datas.Get().([]byte)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}
			msg <- &udpMessage{addr, buf, n}
		}
	}(msg)

	timer := time.Tick(time.Second * 30)
	checkup := time.Tick(time.Second * 30)

	for {
		select {
		case m := <-msg:
			if m.addr != nil && m.data != nil {
				d.HandleMessage(m.addr, m.data[:m.size], t)
				datas.Put(m.data)
			}
		case <-timer:
			if n := d.Route().NumNodes(); n < 1024 {
				d.DoTimer(time.Minute*15, time.Minute*15, time.Hour*6, time.Minute*5)
			}
		case <-checkup:
			if n := d.Route().NumNodes(); n < 1024 {
				d.FindNode(d.ID())
			}
		case <-exit:
			return
		default:
		}
	}
}
