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

type DHTListener struct {
}

func newDHTListener() dht.Listener {
	return &DHTListener{}
}

var tid2 int64
var tors2 string

func (l *DHTListener) GetPeers(id *dht.ID, tor *dht.ID, peers []string) {
	tid2++
	/*
		s := fmt.Sprintln("gp", tid2, tor)
		fmt.Print(s)
		tors2 = s + tors2
	*/
}

var tid int64
var tors string

func (l *DHTListener) AnnouncePeer(id *dht.ID, tor *dht.ID, peer string) {
	tid++
	s := fmt.Sprintln("ap", tid, tor)
	//fmt.Print(s)
	if tid%1000 == 0 {
		tors = ""
	}
	tors = s + tors
}

func (l *DHTListener) OnRequest(meth dht.KadMethod, req *dht.KadRequest) {
	if meth == dht.AnnouncePeer {
		if tid++; tid%1000 == 0 {
			tors = ""
		}
		tor, _ := dht.NewID(req.InfoHash())
		s := fmt.Sprintln("ap", tid, tor)
		tors = s + tors
	}
}

func (l *DHTListener) OnResponse(meth dht.KadMethod, res *dht.KadResponse) {
	if meth == dht.GetPeers {
		tid2++
	}
}

func newDHTServer() (d *dht.DHT, err error) {
	id := dht.NewRandomID()
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return
	}
	handler := newDHTListener()
	d = dht.NewDHT(id, conn.(*net.UDPConn), 16, handler)
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

	var dhts []*dht.DHT
	var w sync.WaitGroup
	for i := 0; i < 1000; i++ {
		d, err := newDHTServer()
		if err != nil {
			continue
		}
		dhts = append(dhts, d)
		w.Add(1)
		go func(d *dht.DHT, idx int, msg chan *udpMessage) {
			defer w.Done()
			conn := d.Conn()
			buf := datas.Get().([]byte)
			for {
				//buf := make([]byte, 1024)
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					fmt.Println(err)
					continue
				}
				//d.HandleMessage(addr, buf[:n])
				msg <- &udpMessage{idx, addr, buf, n}
			}
		}(d, i, msg)
		if err = initDHTServer(d); err != nil {
			fmt.Println(err)
		}
	}

	var numNodes int

	go func() {
		timer := time.Tick(time.Second * 30)
		checkup := time.Tick(time.Second * 30)

		for {
			select {
			case m := <-msg:
				d := dhts[m.idx]
				if m.addr != nil && m.data != nil {
					d.HandleMessage(m.addr, m.data[:m.size])
					datas.Put(m.data)
				}
			case <-timer:
				for _, d := range dhts {
					if n := d.NumNodes(); n < 1024 {
						d.DoTimer(time.Minute*15, time.Minute*15, time.Hour*6)
					}
				}
			case <-checkup:
				var numNodes2 int
				for _, d := range dhts {
					if n := d.NumNodes(); n < 1024 {
						d.FindNode(d.ID())
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
		//fmt.Fprintln(res, dhts[0].Route())
		fmt.Fprintln(res, "---------------------------------------------------")
		res.Write([]byte(tors))
		/*
			fmt.Fprintln(res, "---------------------------------------------------")
			res.Write([]byte(tors2))
		*/
	})
	http.ListenAndServe(":6882", nil)

	w.Wait()
}
