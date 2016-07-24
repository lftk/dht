package dht

// Listener interface
type Listener interface {
	GetPeers(id *ID, tor *ID)
	AnnouncePeer(id *ID, tor *ID, peer *Peer)
}
