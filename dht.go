package dht

import (
	"fmt"
	"net"
	"time"
)

// DHT server
type DHT struct {
	cfg     *Config
	buf     []byte
	exit    chan bool
	conn    *net.UDPConn
	route   *Table
	handler Handler
	storage Storage
	secret  *Secret
}

// NewDHT returns DHT
func NewDHT(cfg *Config) *DHT {
	d := &DHT{
		cfg:    cfg,
		buf:    make([]byte, 4096),
		exit:   make(chan bool),
		secret: NewSecret(),
	}
	if cfg == nil {
		d.cfg = NewConfig()
	}
	d.route = NewTable(d.ID())
	return d
}

// Run dht server
func (d *DHT) Run(handler Handler) (err error) {
	conn, err := net.ListenPacket("udp", d.cfg.Address)
	if err != nil {
		return
	}
	defer conn.Close()
	d.conn = conn.(*net.UDPConn)
	d.handler = handler
	d.loop()
	return
}

// Exit dht server
func (d *DHT) Exit() {
	close(d.exit)
}

// ID returns dht id
func (d *DHT) ID() *ID {
	if d.cfg.ID == nil {
		d.cfg.ID = NewRandomID()
	}
	return d.cfg.ID
}

// Addr returns dht address
func (d *DHT) Addr() *net.UDPAddr {
	if d.conn != nil {
		return d.conn.LocalAddr().(*net.UDPAddr)
	}
	return nil
}

// Route returns route table
func (d *DHT) Route() *Table {
	return d.route
}

func (d *DHT) loop() {
	msg := make(chan *udpMessage, 1024)
	go d.recvMessage(msg)

	d.initialize()

	cleanup := time.Tick(30 * time.Second)
	secret := time.Tick(15 * time.Minute)

	for {
		select {
		case <-d.exit:
			goto EXIT
		case <-cleanup:
			d.cleanup()
		case <-secret:
			d.secret.Update()
		case m := <-msg:
			d.handleMessage(m)
		default:
		}
	}

EXIT:
	d.unInitialize()
}

func (d *DHT) initialize() {
	data := map[string]interface{}{
		"id":     d.ID().Bytes(),
		"target": d.ID().Bytes(),
	}
	tid := encodeTID("find_node", 0)
	msg := NewQueryMessage(tid, "find_node", data)
	b, _ := EncodeMessage(msg)
	for _, route := range d.cfg.Routes {
		addr, err := net.ResolveUDPAddr("udp", route)
		if err == nil {
			d.conn.WriteToUDP(b, addr)
		}
	}

	if d.handler != nil {
		d.handler.Initialize()
	}
}

func (d *DHT) unInitialize() {
	if d.handler != nil {
		d.handler.UnInitialize()
	}
}

func (d *DHT) cleanup() {
	d.route.Map(func(b *Bucket) bool {
		if b.IsGood() {
			d.cleanupBucket(b)
		} else {
			d.findNode(b.RandomID())
		}
		return true
	})
}

func (d *DHT) cleanupBucket(b *Bucket) {
	b.Map(func(n *Node) bool {
		if n.IsGood() == false {
			d.ping(n)
		}
		return true
	})
}

func (d *DHT) handleMessage(msg *udpMessage) (err error) {
	var h KadMsgHeader
	if err = DecodeMessage(msg.data, &h); err != nil {
		return
	}
	switch h.Type() {
	case QueryMessage:
		var req KadRequest
		if err = DecodeMessage(msg.data, &req); err != nil {
			return
		}
		d.handleQueryMessage(h.TID(), msg.addr, &req)
	case ReplyMessage:
		var res KadResponse
		if err = DecodeMessage(msg.data, &res); err != nil {
			return
		}
		d.handleReplyMessage([]byte(h.TID()), msg.addr, &res)
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

	fmt.Println("[query]", req.Method, id, addr)

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
	n := d.find(res.ID())
	if n != nil {
		n.Update()
	} else {
		d.insertNode(res.ID(), addr)
	}

	q, id := decodeTID(tid)

	fmt.Println("[reply]", q, id, res.ID(), addr)

	switch q {
	case "ping":
	case "find_node":
		d.handleFindNode(res.Nodes())
	case "get_peers":
		d.handleGetPeers(res.Values(), res.Nodes())
	case "announce_peer":
		_ = id
	}
}

func (d *DHT) handlePing(res *KadResponse) {
}

func (d *DHT) handleFindNode(nodes []byte) {
	for k, v := range DecodeCompactNode(nodes) {
		id, _ := NewID(k[:])
		addr, _ := net.ResolveUDPAddr("udp", v)
		d.insertNode(id, addr)
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
		for k, v := range DecodeCompactNode(nodes) {
			id, _ := NewID(k[:])
			addr, _ := net.ResolveUDPAddr("udp", v)
			d.insertNode(id, addr)
			// insert search node
		}
	}
}

func (d *DHT) handleAnnouncePeer() {
}

func encodeTID(q string, id uint16) (b []byte) {
	b = make([]byte, 4)
	switch q {
	case "ping":
		b[0] = 'p'
		b[1] = 'n'
	case "find_node":
		b[0] = 'f'
		b[1] = 'n'
	case "get_peers":
		b[0] = 'g'
		b[1] = 'p'
	case "announce_peer":
		b[0] = 'a'
		b[1] = 'p'
	}
	if id != 0 {
		b[2] = byte(id & 0xFF00 >> 8)
		b[3] = byte(id & 0x00FF)
	}
	return
}

func decodeTID(tid []byte) (q string, id uint16) {
	if len(tid) == 4 {
		id = (uint16(tid[2]) << 8) | uint16(tid[3])
		if tid[0] == 'p' && tid[1] == 'n' {
			q = "ping"
		} else if tid[0] == 'f' && tid[1] == 'n' {
			q = "find_node"
		} else if tid[0] == 'g' && tid[1] == 'p' {
			q = "get_peers"
		} else if tid[0] == 'a' && tid[1] == 'p' {
			q = "announce_peer"
		}
	}
	return
}

func (d *DHT) sendQueryMessage(nodes []*Node, q string, id uint16, data map[string]interface{}) {
	tid := encodeTID(q, id)
	msg := NewQueryMessage(tid, q, data)
	d.sendMessage(nodes, msg)
}

func (d *DHT) sendReplyMessage(addrs []*net.UDPAddr, tid []byte, data map[string]interface{}) {
	msg := NewReplyMessage(tid, data)
	if b, err := EncodeMessage(msg); err == nil {
		for _, addr := range addrs {
			d.conn.WriteToUDP(b, addr)
		}
	}
}

func (d *DHT) ping(n *Node) {
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.sendQueryMessage([]*Node{n}, "ping", 0, data)
}

func (d *DHT) findNode(id *ID) {
	data := map[string]interface{}{
		"id":     d.ID().Bytes(),
		"target": id.Bytes(),
	}
	d.sendQueryMessage(d.route.Lookup(id), "find_node", 0, data)
}

func (d *DHT) getPeers(id *ID) {
	data := map[string]interface{}{
		"id":        d.ID().Bytes(),
		"info_hash": id.Bytes(),
	}
	d.sendQueryMessage(d.route.Lookup(id), "get_peers", 0, data)
}

func (d *DHT) sendMessage(nodes []*Node, msg interface{}) {
	if b, err := EncodeMessage(msg); err == nil {
		for _, node := range nodes {
			d.conn.WriteToUDP(b, node.Addr())
		}
	}
}

type udpMessage struct {
	addr *net.UDPAddr
	data []byte
}

func (d *DHT) recvMessage(msg chan *udpMessage) {
	for {
		buf := make([]byte, d.cfg.PacketSize)
		n, addr, err := d.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		msg <- &udpMessage{addr, buf[:n]}

		select {
		case <-d.exit:
			return
		default:
		}
	}
}

func (d *DHT) announcePeer() {
}

func (d *DHT) replyPing(tid []byte, addr *net.UDPAddr) {
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.sendReplyMessage([]*net.UDPAddr{addr}, tid, data)
}

func (d *DHT) replyFindNode(tid []byte, addr *net.UDPAddr, target *ID) {
	nodes := d.route.Lookup(target)
	data := map[string]interface{}{
		"id":    d.ID().Bytes(),
		"nodes": EncodeCompactNode(nodes),
	}
	d.sendReplyMessage([]*net.UDPAddr{addr}, tid, data)
}

func (d *DHT) replyGetPeers(tid []byte, addr *net.UDPAddr, tor *ID) {
	data := map[string]interface{}{
		"id":    d.ID().Bytes(),
		"token": d.secret.Create(addr.String()),
	}
	if peers := d.storage.GetPeers(tor); peers != nil {
		data["values"] = nil
	} else {
		nodes := d.route.Lookup(tor)
		data["nodes"] = EncodeCompactNode(nodes)
	}
	d.sendReplyMessage([]*net.UDPAddr{addr}, tid, data)
}

func (d *DHT) replyAnnouncePeer(tid []byte, addr *net.UDPAddr, req *KadRequest) {
	b := d.secret.Match(addr.String(), req.Token())
	if b == false {
		// send error message
		return
	}
	data := map[string]interface{}{
		"id": d.ID().Bytes(),
	}
	d.sendReplyMessage([]*net.UDPAddr{addr}, tid, data)
}

func (d *DHT) insertNode(id *ID, addr *net.UDPAddr) *Node {
	n := NewNode(id, addr)
	d.route.Append(n)
	return n
}

func (d *DHT) find(id *ID) (n *Node) {
	if b := d.route.Find(id); b != nil {
		n = b.Find(id)
	}
	return
}

func (d *DHT) insertOrUpdate(id *ID, addr *net.UDPAddr) {
	if b := d.route.Find(id); b != nil {
		if n := b.Find(id); n != nil {
			n.Update()
		} else {
			b.Insert(id, addr)
		}
		b.Update()
	}
}
