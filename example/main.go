package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/4396/dht"
)

func resolveAddr(b []byte) string {
	ip, port := dht.ResolveAddr(b)
	return fmt.Sprintf("%s:%d", ip, port)
}

type dhtQueryTracker struct {
	d *dht.DHT
}

func (t *dhtQueryTracker) Ping(id *dht.ID) {
}

func (t *dhtQueryTracker) FindNode(id *dht.ID, target *dht.ID) {
}

var tid2 int64
var tors2 string

func (t *dhtQueryTracker) GetPeers(id *dht.ID, tor *dht.ID) {
	tid2++
}

var tid int64
var tors string

func (t *dhtQueryTracker) AnnouncePeer(id *dht.ID, tor *dht.ID, peer []byte) {
	if tid++; tid%1000 == 0 {
		tors = ""
	}
	s := fmt.Sprintln("ap", tid, tor)
	tors = s + tors

	fmt.Println("ap", string(peer), resolveAddr(peer))
}

type dhtReplyTracker struct {
	d *dht.DHT
}

func (t *dhtReplyTracker) Ping(id *dht.ID) {
}

func (t *dhtReplyTracker) FindNode(id *dht.ID, nodes []byte) {
}

func (t *dhtReplyTracker) GetPeers(id *dht.ID, peers [][]byte, nodes []byte) {
	fmt.Println("----GetPeers")
	for _, p := range peers {
		fmt.Println("gp", string(p), resolveAddr(p))
	}
}

func (t *dhtReplyTracker) AnnouncePeer(id *dht.ID) {
}

type dhtErrorTracker struct {
	d *dht.DHT
}

func (t *dhtErrorTracker) Error(code int, msg string) {
	//fmt.Println(code, msg)
}

type dhtServer struct {
	d *dht.DHT
	t *dht.Tracker
}

func newDHTServer() (s *dhtServer, err error) {
	id := dht.NewRandomID()
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return
	}
	d := dht.NewDHT(id, conn.(*net.UDPConn), 16)
	t := dht.NewTracker(
		&dhtQueryTracker{d},
		&dhtReplyTracker{d},
		&dhtErrorTracker{d},
	)
	s = &dhtServer{d, t}
	return
}

func dhtNodeNums(d *dht.DHT) (n int) {
	d.Route().Map(func(b *dht.Bucket) bool {
		n += b.Count()
		return true
	})
	return
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
	idx  int
	addr *net.UDPAddr
	data []byte
	size int
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	msg := make(chan *udpMessage, 1024)

	datas := &sync.Pool{New: func() interface{} {
		return make([]byte, 1024)
	}}

	var svrs []*dhtServer
	var w sync.WaitGroup
	for i := 0; i < 1000; i++ {
		s, err := newDHTServer()
		if err != nil {
			continue
		}
		svrs = append(svrs, s)
		w.Add(1)
		go func(d *dht.DHT, idx int, msg chan *udpMessage) {
			defer w.Done()

			if err = initDHTServer(d); err != nil {
				fmt.Println(err)
			}

			conn := d.Conn()
			buf := datas.Get().([]byte)
			for {
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					fmt.Println(err)
					continue
				}
				msg <- &udpMessage{idx, addr, buf, n}
			}
		}(s.d, i, msg)
	}

	var numNodes int

	go func() {
		timer := time.Tick(time.Second * 30)
		checkup := time.Tick(time.Second * 30)

		for {
			select {
			case m := <-msg:
				s := svrs[m.idx]
				if m.addr != nil && m.data != nil {
					s.d.HandleMessage(m.addr, m.data[:m.size], s.t)
					datas.Put(m.data)
				}
			case <-timer:
				for _, s := range svrs {
					if n := s.d.Route().NumNodes(); n < 1024 {
						s.d.DoTimer(time.Minute*15, time.Minute*15, time.Hour*6)
					}
				}
			case <-checkup:
				var numNodes2 int
				tor, _ := dht.ResolveID("004aa73f1a3001fb6ecf545336f155123aee4941")
				for _, s := range svrs {
					if n := s.d.Route().NumNodes(); n < 1024 {
						s.d.FindNode(s.d.ID())
						s.d.GetPeers(tor)
						numNodes2 += n
					}
				}
				numNodes = numNodes2
				fmt.Println(numNodes, tid, tid2)
			default:
			}
		}
	}()

	startupTime := time.Now()
	go http.HandleFunc("/dht", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(res, "===", startupTime, numNodes, tid, tid2)
		fmt.Fprintln(res, "---------------------------------------------------")
		res.Write([]byte(tors))
	})
	http.ListenAndServe(":6882", nil)

	w.Wait()
}
