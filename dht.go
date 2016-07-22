package dht

import (
	"fmt"
	"math"
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

func NewDHT2(id *ID, conn *net.UDPConn, handler Handler) *DHT {
	d := &DHT{
		conn:    conn,
		exit:    make(chan bool),
		secret:  NewSecret(),
		handler: handler,
	}
	d.route = NewTable(id)
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
	return d.route.id
}

// Conn returns dht connection
func (d *DHT) Conn() *net.UDPConn {
	return d.conn
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
			if node := b.Random(); node != nil {
				d.findNode(node.ID())
			}
		}
		return true
	})

	if d.handler != nil {
		d.handler.Cleanup()
	}
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
			fmt.Println(string(msg.data))
			fmt.Println(err)
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

func (d *DHT) sendQueryMessage(nodes []*Node, q string, id int16, data map[string]interface{}) {
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

func (d *DHT) FindNodeFromAddr(id *ID, addr *net.UDPAddr) error {
	data := map[string]interface{}{
		"id":     d.ID().Bytes(),
		"target": id.Bytes(),
	}
	return d.queryMessage("find_node", 0, addr, data)
}

func (d *DHT) FindNode(id *ID) {
	d.findNode(id)
}

func (d *DHT) findNode(id *ID) {
	if addrs := d.lookup(id); addrs != nil {
		data := map[string]interface{}{
			"id":     d.ID().Bytes(),
			"target": id.Bytes(),
		}
		d.batchQueryMessage("find_node", 0, addrs, data)
	}
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

func (d *DHT) sendMessage(nodes []*Node, msg interface{}) {
	if b, err := EncodeMessage(msg); err == nil {
		for _, node := range nodes {
			d.conn.WriteToUDP(b, node.Addr())
		}
	}
}

func (d *DHT) HandleMessage(addr *net.UDPAddr, data []byte) error {
	return d.handleMessage(&udpMessage{addr, data})
}

func (d *DHT) Cleanup() {
	d.cleanup()
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
		"token": d.secret.Create(addr.String()),
	}
	if peers := d.storage.GetPeers(tor); peers != nil {
		data["values"] = nil
	} else if nodes := d.route.Lookup(tor); nodes != nil {
		data["nodes"] = encodeCompactNodes(nodes)
	}
	d.replyMessage(tid, addr, data)

	if d.handler != nil {
		d.handler.GetPeers(tor)
	}
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
	d.replyMessage(tid, addr, data)

	if d.handler != nil {
		d.handler.AnnouncePeer(req.InfoHash(), nil)
	}
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
			_, err := d.route.Insert(id, addr)
			if err != nil {
				//fmt.Println(err, id, addr)
			}
		}
		b.Update()
	}
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

func (d *DHT) sendMsg(addr *net.UDPAddr, data []byte) (err error) {
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
		err = d.sendMsg(addr, b)
	}
	return
}

func (d *DHT) replyMessage(tid []byte, addr *net.UDPAddr, data map[string]interface{}) (err error) {
	msg := NewReplyMessage(tid, data)
	if b, err := encodeMessage(msg); err == nil {
		err = d.sendMsg(addr, b)
	}
	return
}

func (d *DHT) batchQueryMessage(q string, no int16, addrs []*net.UDPAddr, data map[string]interface{}) (n int, err error) {
	msg := NewQueryMessage(encodeTID(q, no), q, data)
	if b, err := encodeMessage(msg); err == nil {
		for _, addr := range addrs {
			err = d.sendMsg(addr, b)
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
		if id < 0 {
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
