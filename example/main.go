package main

import (
	"log"

	"github.com/4396/dht"
)

func main() {
	d := dht.NewDHT(nil)
	log.Fatal(d.Run(":0", nil))
}
