package dht

// Handler interface
type Handler interface {
	Initialize()
	UnInitialize()

	Cleanup()
	GetPeers(tor *ID)
	AnnouncePeer(tor *ID, peer *Peer)
}

// Listener interface
type Listener interface {
	GetPeers(id *ID, tor *ID)
	AnnouncePeer(id *ID, tor *ID, peer *Peer)
}
