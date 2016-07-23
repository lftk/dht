package dht

import (
	"math"
	"net"
	"time"
)

// DHT server
type DHT struct {
	conn     *net.UDPConn
	route    *Table
	secret   *Secret
	storage  Storage
	listener Listener
}

// NewDHT returns DHT
func NewDHT(id *ID, conn *net.UDPConn, ksize int, listener Listener) *DHT {
	return &DHT{
		conn:     conn,
		route:    NewTable(id, ksize),
		secret:   NewSecret(),
		listener: listener,
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

// Listener returns listener
func (d *DHT) Listener() Listener {
	return d.listener
}

// NumNodes returns all node count
func (d *DHT) NumNodes() (n int) {
	d.route.Map(func(b *Bucket) bool {
		n += b.Count()
		return true
	})
	return
}

// UpdateSecret update secret
func (d *DHT) UpdateSecret() {
	d.secret.Update()
}

// CleanNodes clean timeout nodes
func (d *DHT) CleanNodes(tm time.Duration) {
	d.route.Map(func(b *Bucket) bool {
		if time.Since(b.time) > tm {
			if n := b.Random(); n != nil {
				d.FindNode(n.ID())
			}
		} else {
			b.clean(func(n *Node) bool {
				if n.pinged > 1 {
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

// HandleMessage handle udp packet
func (d *DHT) HandleMessage(addr *net.UDPAddr, data []byte) (err error) {
	var h KadMsgHeader
	if err = decodeMessage(data, &h); err != nil {
		return
	}
	switch h.Type() {
	case QueryMessage:
		var req KadRequest
		if err = decodeMessage(data, &req); err == nil {
			d.handleQueryMessage(h.TID(), addr, &req)
		}
	case ReplyMessage:
		var res KadResponse
		if err = decodeMessage(data, &res); err == nil {
			d.handleReplyMessage(h.TID(), addr, &res)
		}
	case ErrorMessage:
		// ...
	}
	return
}

func (d *DHT) handleQueryMessage(tid []byte, addr *net.UDPAddr, req *KadRequest) {
	id, err := NewID(req.ID())
	if err != nil {
		return
	}
	d.insertOrUpdate(id, addr)

	switch req.Method {
	case "ping":
		d.replyPing(tid, addr)
	case "find_node":
		d.replyFindNode(tid, addr, req.Target())
	case "get_peers":
		d.replyGetPeers(tid, addr, req.InfoHash())
	case "announce_peer":
		d.replyAnnouncePeer(tid, addr, req)
	}
}

func (d *DHT) handleReplyMessage(tid []byte, addr *net.UDPAddr, res *KadResponse) {
	id, err := NewID(res.ID())
	if err != nil {
		return
	}
	d.insertOrUpdate(id, addr)

	q, no := decodeTID(tid)

	switch q {
	case "ping":
	case "find_node":
		d.handleFindNode(res.Nodes())
	case "get_peers":
		d.handleGetPeers(res.Values(), res.Nodes())
	case "announce_peer":
		_ = no
	default:
		//fmt.Println(string(tid), len(tid))
	}
}

func (d *DHT) handlePing(res *KadResponse) {
}

func (d *DHT) handleFindNode(nodes []byte) {
	for id, addr := range DecodeCompactNode(nodes) {
		d.insertOrUpdate(id, addr)
	}
}

func (d *DHT) handleGetPeers(values []string, nodes []byte) {
	if len(values) > 0 {
		for _, v := range DecodeCompactPeer(values) {
			addr, _ := net.ResolveTCPAddr("tcp", v)
			peer := NewPeer(addr)
			_ = peer
			// insert peer node
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

func (d *DHT) getPeers(id *ID) {
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

func (d *DHT) replyPing(tid []byte, addr *net.UDPAddr) {
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.replyMessage(tid, addr, data)
}

func (d *DHT) replyFindNode(tid []byte, addr *net.UDPAddr, target *ID) {
	if nodes := d.route.Lookup(target); nodes != nil {
		data := map[string]interface{}{
			"id":    d.ID().Bytes(),
			"nodes": encodeCompactNodes(nodes),
		}
		d.replyMessage(tid, addr, data)
	}
}

func (d *DHT) replyGetPeers(tid []byte, addr *net.UDPAddr, tor *ID) {
	data := map[string]interface{}{
		"id":    d.ID().Bytes(),
		"token": d.createToken(addr),
	}
	if peers := d.storage.GetPeers(tor); peers != nil {
		data["values"] = nil
	} else if nodes := d.route.Lookup(tor); nodes != nil {
		data["nodes"] = encodeCompactNodes(nodes)
	}
	d.replyMessage(tid, addr, data)

	if d.listener != nil {
		d.listener.GetPeers(nil, tor)
	}
}

func (d *DHT) replyAnnouncePeer(tid []byte, addr *net.UDPAddr, req *KadRequest) {
	if d.matchToken(addr, req.Token()) == false {
		// send error message
		return
	}
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.replyMessage(tid, addr, data)

	if d.listener != nil {
		d.listener.AnnouncePeer(nil, req.InfoHash(), nil)
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
