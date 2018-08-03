package rust

import (
	"log"
	"net"
	"time"

	"github.com/alliedmodders/blaster/valve"
)

const logSymbol = "👁️‍🗨️ "

// A ServerConfig is the connection information for a Rust server
type ServerConfig struct {
	Hostname string
	Port     int
}

// A PlayerInfo contains information about player counts on a Rust server
type PlayerInfo struct {
	Players      uint8
	MaxPlayers   uint8
	PlayersDelta int8
}

// A Server queries and contains information about a Rust server
type Server struct {
	ServerConfig
	tcpAddr    net.TCPAddr
	Name       string
	PlayerInfo PlayerInfo
}

// NewServer creats a new Server for observing a Rust server
func NewServer(server *ServerConfig) (*Server, error) {
	var sq = Server{}
	rustIP, err := net.ResolveIPAddr("ip", server.Hostname)
	if err != nil {
		log.Println(logSymbol + err.Error())
		return nil, err
	}
	sq.tcpAddr = net.TCPAddr{IP: rustIP.IP, Port: server.Port}

	return &sq, nil
}

// Update queries a Rust server and updates Server with it's new information
func (sq *Server) Update() error {
	query, err := valve.NewServerQuerier(sq.tcpAddr.String(), time.Second*3)
	if err != nil {
		sq.PlayerInfo = PlayerInfo{}
		return err
	}
	defer query.Close()

	info, err := query.QueryInfo()
	if err != nil {
		sq.PlayerInfo = PlayerInfo{}
		return err
	}

	sq.Name = info.Name
	sq.PlayerInfo.PlayersDelta = (int8)(info.Players - sq.PlayerInfo.Players)
	sq.PlayerInfo.MaxPlayers = info.MaxPlayers
	sq.PlayerInfo.Players = info.Players
	return nil
}
