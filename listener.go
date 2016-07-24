package dht

// Listener interface
type Listener interface {
	GetPeers(id *ID, tor *ID, peers []string)
	AnnouncePeer(id *ID, tor *ID, peer string)

	OnRequest(meth KadMethod, req *KadRequest)
	OnResponse(meth KadMethod, res *KadResponse)
}
