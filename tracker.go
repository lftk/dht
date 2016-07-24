package dht

// QueryTracker interface
type QueryTracker interface {
	Ping(id *ID)
	FindNode(id *ID, target *ID)
	GetPeers(id *ID, tor *ID)
	AnnouncePeer(id *ID, tor *ID, peer string)
}

// ReplyTracker interface
type ReplyTracker interface {
	Ping(id *ID)
	FindNode(id *ID, nodes []byte)
	GetPeers(id *ID, peers []string, nodes []byte)
	AnnouncePeer(id *ID)
}

// ErrorTracker interface
type ErrorTracker interface {
	Error(code int, msg string)
}

// Tracker struct
type Tracker struct {
	q QueryTracker
	r ReplyTracker
	e ErrorTracker
}

// NewTracker returns tracker
func NewTracker(q QueryTracker, r ReplyTracker, e ErrorTracker) *Tracker {
	return &Tracker{q, r, e}
}
