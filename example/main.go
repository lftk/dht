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

type DHTHandler struct {
	d *dht.DHT
}

func NewDHTHandler(d *dht.DHT) dht.Handler {
	return &DHTHandler{
		d: d,
	}
}

func (l *DHTHandler) Initialize() {
	//	fmt.Println(l.d.ID())
	//	fmt.Println(l.d.Addr())
}

func (l *DHTHandler) UnInitialize() {
	fmt.Println("--exit--")

	l.d.Route().Map(func(b *dht.Bucket) bool {
		b.Map(func(n *dht.Node) bool {
			return true
		})
		return true
	})
}

func (l *DHTHandler) Cleanup() {
	if n := dhtNodeNums(l.d); n < 1024 {
		l.d.FindNode(l.d.ID())
	}
}

var tid2 int
var tors2 string

func (l *DHTHandler) GetPeers(tor *dht.ID) {
	/*
		tid2++
		s := fmt.Sprintln("gp", tid2, tor)
		fmt.Print(s)
		tors2 = s + tors2
	*/
}

var tid int64
var tors string

func (l *DHTHandler) AnnouncePeer(tor *dht.ID, peer *dht.Peer) {
	tid++
	s := fmt.Sprintln("ap", tid, tor)
	fmt.Print(s)
	if tid%1000 == 0 {
		tors = ""
	}
	tors = s + tors
}

/*
func (l *DHTHandler) HandleError(e *dht.KadError) {
	fmt.Println("err:", e.Value(), e.String())
}
*/

func newDHTServer() (d *dht.DHT, err error) {
	id := dht.NewRandomID()
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return
	}
	handler := NewDHTHandler(nil)
	d = dht.NewDHT2(id, conn.(*net.UDPConn), handler)
	handler.(*DHTHandler).d = d
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
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	msg := make(chan *udpMessage, 10240)
	var dhts []*dht.DHT
	var w sync.WaitGroup
	for i := 0; i < /*runtime.NumCPU()*/ 1000; i++ {
		d, err := newDHTServer()
		if err != nil {
			continue
		}
		dhts = append(dhts, d)
		w.Add(1)
		go func(d *dht.DHT, idx int, msg chan *udpMessage) {
			defer w.Done()
			conn := d.Conn()
			for {
				buf := make([]byte, 1024)
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					fmt.Println(err)
					continue
				}
				msg <- &udpMessage{idx, addr, buf[:n]}
			}
		}(d, i, msg)
		if err = initDHTServer(d); err != nil {
			fmt.Println(err)
		}
	}

	go func() {
		cleanup := time.Tick(time.Second * 30)
		for {
			select {
			case m := <-msg:
				d := dhts[m.idx]
				if m.addr != nil && m.data != nil {
					d.HandleMessage(m.addr, m.data)
				}
			case <-cleanup:
				for _, d := range dhts {
					d.Cleanup()
				}
			default:
			}
		}
	}()

	go http.HandleFunc("/dht", func(res http.ResponseWriter, req *http.Request) {
		var count int
		for _, d := range dhts {
			//s := fmt.Sprintf("%02d %s %04d\n", i, d.ID(), dhtNodeNums(d))
			//res.Write([]byte(s))
			count += dhtNodeNums(d)
		}
		fmt.Println("//", count, tid, tid2)

		fmt.Fprintln(res, "===", count)
		//fmt.Fprintln(res, dhts[0].Route())
		fmt.Fprintln(res, "---------------------------------------------------")
		res.Write([]byte(tors))
		/*
			fmt.Fprintln(res, "---------------------------------------------------")
			res.Write([]byte(tors2))
		*/
	})
	go http.ListenAndServe(":6882", nil)

	w.Wait()
}
