package dht

// Handler interface
type Handler interface {
	Initialize()
	UnInitialize()

	Cleanup()
	GetPeers(tor *ID)
	AnnouncePeer(tor *ID, peer *Peer)
}
