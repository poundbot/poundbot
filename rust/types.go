package rust

// A ServerConfig is the connection information for a Rust server
type ServerConfig struct {
	Hostname string
	Port     int
}

// A PlayerInfo contains information about player counts on a Rust server and
// methods to read it's state
type PlayerInfo struct {
	Players      uint8
	MaxPlayers   uint8
	PlayersDelta int8
}
