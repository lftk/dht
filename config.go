package dht

// Config infomation
type Config struct {
	ID         *ID
	Address    string
	PacketSize int
	Routes     []string
	MinNodes   int
}

// NewConfig returns config
func NewConfig() *Config {
	cfg := &Config{
		Address:    ":0",
		PacketSize: 8192,
		Routes:     make([]string, 0),
		MinNodes:   -1,
	}
	cfg.Routes = append(cfg.Routes, "router.magnets.im:6881")
	cfg.Routes = append(cfg.Routes, "router.bittorrent.com:6881")
	cfg.Routes = append(cfg.Routes, "dht.transmissionbt.com:6881")
	cfg.Routes = append(cfg.Routes, "router.utorrent.com:6881")
	return cfg
}
