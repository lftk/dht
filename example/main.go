package main

import (
	"fmt"
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
	fmt.Println(l.d.ID())
	fmt.Println(l.d.Addr())
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var w sync.WaitGroup
	w.Add(1)
	for i := 0; i < 1; i++ {
		go func() {
			defer w.Done()
			dhtServer()
		}()
	}
	w.Wait()
}
