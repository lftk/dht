package dht

import (
	"errors"
	"fmt"
	"math"
	"net"
	"time"
)

// DHT server
type DHT struct {
	conn     *net.UDPConn
	route    *Table
	secret   *Secret
	storages *storages
	tsecret  time.Time
}

// NewDHT returns DHT
func NewDHT(id *ID, conn *net.UDPConn, ksize int) *DHT {
	return &DHT{
		conn:     conn,
		route:    NewTable(id, ksize),
		secret:   NewSecret(),
		storages: newStorages(),
	}
}

// ID returns dht id
func (d *DHT) ID() *ID {
	return d.route.id
}

// Conn returns dht connection
func (d *DHT) Conn() *net.UDPConn {
	return d.conn
}

// Addr returns dht address
func (d *DHT) Addr() *net.UDPAddr {
	if conn := d.conn; conn != nil {
		return conn.LocalAddr().(*net.UDPAddr)
	}
	return nil
}

// Route returns route table
func (d *DHT) Route() *Table {
	return d.route
}

// NumNodes returns all node count
func (d *DHT) NumNodes() (n int) {
	d.route.Map(func(b *Bucket) bool {
		n += b.Count()
		return true
	})
	return
}

func (d *DHT) cleanNodes(tm time.Duration) {
	d.route.Map(func(b *Bucket) bool {
		if time.Since(b.time) > tm {
			if n := b.Random(); n != nil {
				d.FindNode(n.ID())
			}
		} else {
			b.clean(func(n *Node) bool {
				if n.pinged > 0 {
					return true
				}
				if time.Since(n.time) > tm {
					d.Ping(n.addr)
					n.pinged++
				}
				return false
			})
		}
		return true
	})
}

func (d *DHT) cleanPeers(tm time.Duration) {

}

// DoTimer update secret, clean nodes and peers
func (d *DHT) DoTimer(secret, node, peer time.Duration) {
	if time.Since(d.tsecret) >= secret {
		d.tsecret = time.Now()
		d.secret.Update()
	}
	d.cleanNodes(node)
	d.cleanPeers(peer)
}

// HandleMessage handle udp packet
func (d *DHT) HandleMessage(addr *net.UDPAddr, data []byte, t *Tracker) (err error) {
	var msg KadMessage
	err = decodeMessage(data, &msg)
	if err != nil {
		return
	}
	switch msg.Y {
	case "q":
		err = d.handleQueryMessage(addr, msg.T, msg.Q, &msg.A, t.q)
	case "r":
		err = d.handleReplyMessage(addr, msg.T, &msg.R, t.r)
	case "e":
		err = d.handleErrorMessage(msg.E, t.e)
	}
	return
}

func (d *DHT) handleErrorMessage(err []interface{}, t ErrorTracker) error {
	if len(err) == 2 {
		val, ok0 := err[0].(int64)
		str, ok1 := err[1].(string)
		if ok0 && ok1 {
			if t != nil {
				t.Error(int(val), str)
			}
			return nil
		}
	}
	return errors.New("Is not a standard error message")
}

func (d *DHT) handleQueryMessage(addr *net.UDPAddr, tid []byte, meth string, args *KadArguments, t QueryTracker) (err error) {
	id, err := NewID(args.ID)
	if err != nil {
		return
	}
	d.insertOrUpdate(id, addr)

	switch meth {
	case "ping":
		d.replyPing(addr, tid)
		if t != nil {
			t.Ping(id)
		}
	case "find_node":
		target, err := NewID(args.Target)
		if err == nil {
			d.replyFindNode(addr, tid, target)
			if t != nil {
				t.FindNode(id, target)
			}
		}
	case "get_peers":
		tor, err := NewID(args.InfoHash)
		if err == nil {
			d.replyGetPeers(addr, tid, tor)
			if t != nil {
				t.GetPeers(id, tor)
			}
		}
	case "announce_peer":
		tor, err := NewID(args.InfoHash)
		if err == nil {
			peer := fmt.Sprintf("%s:%d", addr.IP.String(), args.Port)
			d.replyAnnouncePeer(addr, tid, args.Token, tor, peer)
			if t != nil {
				t.AnnouncePeer(id, tor, peer)
			}
		}
	}
	return
}

func (d *DHT) handleReplyMessage(addr *net.UDPAddr, tid []byte, resp *KadResponse, t ReplyTracker) (err error) {
	id, err := NewID(resp.ID)
	if err != nil {
		return
	}
	d.insertOrUpdate(id, addr)

	q, no := decodeTID(tid)

	switch q {
	case "ping":
		if t != nil {
			t.Ping(id)
		}
	case "find_node":
		d.handleFindNode(resp.Nodes)
		if t != nil {
			t.FindNode(id, nil)
		}
	case "get_peers":
		var tor *ID
		d.handleGetPeers(tor, resp.Values, resp.Nodes)
		if t != nil {
			t.GetPeers(id, resp.Values, resp.Nodes)
		}
	case "announce_peer":
		_ = no
	}
	return
}

func (d *DHT) handlePing() {
}

func (d *DHT) handleFindNode(nodes []byte) {
	for id, addr := range DecodeCompactNode(nodes) {
		d.insertOrUpdate(id, addr)
	}
}

func (d *DHT) handleGetPeers(tor *ID, values []string, nodes []byte) {
	if tor == nil {
		return
	}
	if len(values) > 0 {
		for _, peer := range values {
			d.storePeer(tor, peer)
		}
	} else if len(nodes) > 0 {
		for id, addr := range DecodeCompactNode(nodes) {
			d.insertOrUpdate(id, addr)
		}
	}
}

func (d *DHT) handleAnnouncePeer() {
}

// Ping a address
func (d *DHT) Ping(addr *net.UDPAddr) error {
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	return d.queryMessage("ping", 0, addr, data)
}

func (d *DHT) ping(n *Node) {
	d.Ping(n.Addr())
}

// FindNodeFromAddr find node from address
func (d *DHT) FindNodeFromAddr(id *ID, addr *net.UDPAddr) error {
	data := map[string]interface{}{
		"id":     d.ID().Bytes(),
		"target": id.Bytes(),
	}
	return d.queryMessage("find_node", 0, addr, data)
}

// FindNodeFromAddrs find node from some address
func (d *DHT) FindNodeFromAddrs(id *ID, addrs []*net.UDPAddr) (int, error) {
	data := map[string]interface{}{
		"id":     d.ID().Bytes(),
		"target": id.Bytes(),
	}
	return d.batchQueryMessage("find_node", 0, addrs, data)
}

// FindNode find node
func (d *DHT) FindNode(id *ID) (err error) {
	if addrs := d.lookup(id); addrs != nil {
		data := map[string]interface{}{
			"id":     d.ID().Bytes(),
			"target": id.Bytes(),
		}
		d.batchQueryMessage("find_node", 0, addrs, data)
	}
	return
}

func (d *DHT) findNode(id *ID) {
	d.findNode(id)
}

// GetPeers search info hash
func (d *DHT) GetPeers(id *ID) {
	if addrs := d.lookup(id); addrs != nil {
		data := map[string]interface{}{
			"id":        d.ID().Bytes(),
			"info_hash": id.Bytes(),
		}
		d.batchQueryMessage("get_peers", 0, addrs, data)
	}
}

func (d *DHT) announcePeer() {
}

func (d *DHT) replyPing(addr *net.UDPAddr, tid []byte) {
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.replyMessage(tid, addr, data)
}

func (d *DHT) replyFindNode(addr *net.UDPAddr, tid []byte, target *ID) {
	if nodes := d.route.Lookup(target); nodes != nil {
		data := map[string]interface{}{
			"id":    d.ID().Bytes(),
			"nodes": encodeCompactNodes(nodes),
		}
		d.replyMessage(tid, addr, data)
	}
}

func (d *DHT) replyGetPeers(addr *net.UDPAddr, tid []byte, tor *ID) {
	data := map[string]interface{}{
		"id":    d.ID().Bytes(),
		"token": d.createToken(addr),
	}
	if peers := d.getPeers(tor); peers != nil {
		data["values"] = peers
	} else if nodes := d.route.Lookup(tor); nodes != nil {
		data["nodes"] = encodeCompactNodes(nodes)
	}
	d.replyMessage(tid, addr, data)
}

func (d *DHT) replyAnnouncePeer(addr *net.UDPAddr, tid []byte, token []byte, tor *ID, peer string) {
	if d.matchToken(addr, token) == false {
		// send error message
		return
	}
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.replyMessage(tid, addr, data)

	err := d.storePeer(tor, peer)
	if err != nil {
		return
	}
}

func (d *DHT) createToken(addr *net.UDPAddr) []byte {
	b := []byte(addr.String())
	return d.secret.Create(b)
}

func (d *DHT) matchToken(addr *net.UDPAddr, token []byte) bool {
	b := []byte(addr.String())
	return d.secret.Match(b, token)
}

func (d *DHT) find(id *ID) (n *Node) {
	if b := d.route.Find(id); b != nil {
		n = b.Find(id)
	}
	return
}

func (d *DHT) insertOrUpdate(id *ID, addr *net.UDPAddr) (n *Node, err error) {
	if b := d.route.Find(id); b != nil {
		if n = b.Find(id); n != nil {
			n.Update()
		} else {
			n, err = d.route.Insert(id, addr)
		}
		b.Update()
	}
	return
}

func (d *DHT) storePeer(tor *ID, peer string) error {
	if d.storages.Count() > 102400 {
		return errors.New("102400")
	}
	s := d.storages.Get(tor)
	if s.Count() > 1024 {
		return errors.New("1024")
	}
	s.Insert(peer)
	return nil
}

func (d *DHT) getPeers(tor *ID) (ps []string) {
	if s := d.storages.Find(tor); s != nil {
		s.Map(func(peer string, time time.Time) bool {
			ps = append(ps, peer)
			return len(ps) < d.route.ksize
		})
	}
	return
}

func encodeCompactNodes(nodes []*Node) []byte {
	infos := make(map[*ID]*net.UDPAddr)
	for _, n := range nodes {
		infos[n.ID()] = n.Addr()
	}
	return EncodeCompactNode(infos)
}

func (d *DHT) lookup(id *ID) (addrs []*net.UDPAddr) {
	if nodes := d.route.Lookup(id); len(nodes) > 0 {
		addrs = make([]*net.UDPAddr, len(nodes))
		for i, node := range nodes {
			addrs[i] = node.Addr()
		}
	}
	return
}

func (d *DHT) sendMessage(addr *net.UDPAddr, data []byte) (err error) {
	for n, nn := 0, 0; nn < len(data); nn += n {
		n, err = d.conn.WriteToUDP(data[nn:], addr)
		if err != nil {
			break
		}
	}
	return
}

func (d *DHT) queryMessage(q string, no int16, addr *net.UDPAddr, data map[string]interface{}) (err error) {
	msg := NewQueryMessage(encodeTID(q, no), q, data)
	if b, err := encodeMessage(msg); err == nil {
		err = d.sendMessage(addr, b)
	}
	return
}

func (d *DHT) replyMessage(tid []byte, addr *net.UDPAddr, data map[string]interface{}) (err error) {
	msg := NewReplyMessage(tid, data)
	if b, err := encodeMessage(msg); err == nil {
		err = d.sendMessage(addr, b)
	}
	return
}

func (d *DHT) batchQueryMessage(q string, no int16, addrs []*net.UDPAddr, data map[string]interface{}) (n int, err error) {
	msg := NewQueryMessage(encodeTID(q, no), q, data)
	if b, err := encodeMessage(msg); err == nil {
		for _, addr := range addrs {
			err = d.sendMessage(addr, b)
			if err != nil {
				break
			}
			n++
		}
	}
	return
}

var (
	tids = map[string]string{
		"ping": "pnpn", "find_node": "fnfn",
		"get_peers": "gpgp", "announce_peer": "apap",
	}
	vals = map[string]string{
		"pn": "ping", "fn": "find_node",
		"gp": "get_peers", "ap": "announce_peer",
	}
)

func encodeTID(q string, id int16) (val []byte) {
	if tid, ok := tids[q]; ok {
		uid := uint16(id)
		if id <= 0 {
			uid = math.MaxUint16
		}
		val = []byte(tid)
		val[2] = byte(uid & 0xFF00 >> 8)
		val[3] = byte(uid & 0x00FF)
	}
	return
}

func decodeTID(tid []byte) (q string, id int16) {
	if len(tid) == 4 {
		if val, ok := vals[string(tid[:2])]; ok {
			uid := (uint16(tid[2]) << 8) | uint16(tid[3])
			if uid != math.MaxUint16 {
				id = int16(uid)
			}
			q = val
		}
	}
	return
}
