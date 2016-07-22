package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"

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
	l.d.FindNode(l.d.ID())
}

var tid2 int
var tors2 string

func (l *DHTHandler) GetPeers(tor *dht.ID) {
	tid2++
	s := fmt.Sprintln(tid2, tor)
	fmt.Print(s)
	tors2 = s + tors2
}

var tid int
var tors string

func (l *DHTHandler) AnnouncePeer(tor *dht.ID, peer *dht.Peer) {
	tid++
	s := fmt.Sprintln(tid, tor)
	fmt.Print(s)
	tors = s + tors
}

/*
func (l *DHTHandler) HandleError(e *dht.KadError) {
	fmt.Println("err:", e.Value(), e.String())
}
*/

func dhtServer() {
	cfg := dht.NewConfig()
	//cfg.Address = ":6881"
	//cfg.ID, _ = dht.ResolveID("7c8e2aab1f3117120450ebde3e9c0bc82bdf0b59")

	d := dht.NewDHT(cfg)
	err := d.Run(NewDHTHandler(d))
	if err != nil {
		fmt.Println(err)
	}
}

func dhtNodeNums(d *dht.DHT) (n int) {
	d.Route().Map(func(b *dht.Bucket) bool {
		n += b.Count()
		return true
	})
	return
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var dhts []*dht.DHT

	var w sync.WaitGroup
	w.Add(1)
	for i := 0; i < runtime.NumCPU(); i++ {
		cfg := dht.NewConfig()
		//cfg.MinNodes = 512
		d := dht.NewDHT(cfg)
		dhts = append(dhts, d)
		go func() {
			defer w.Done()
			d.Run(NewDHTHandler(d))
		}()
	}

	//	cfg.Address = ":6881"
	//	cfg.ID, _ = dht.ResolveID("7c8e2aab1f3117120450ebde3e9c0bc82bdf0b59")

	http.HandleFunc("/dht", func(res http.ResponseWriter, req *http.Request) {
		for i, d := range dhts {
			s := fmt.Sprintf("%02d %s %04d\n", i, d.ID(), dhtNodeNums(d))
			res.Write([]byte(s))
		}
		fmt.Fprintln(res, "---------------------------------------------------")
		res.Write([]byte(tors))
		fmt.Fprintln(res, "---------------------------------------------------")
		res.Write([]byte(tors2))
		return
	})
	go http.ListenAndServe(":6882", nil)

	w.Wait()
}
