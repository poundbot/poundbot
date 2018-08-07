package rust

import (
	"log"
	"net"
	"time"

	"github.com/alliedmodders/blaster/valve"
)

const qLogSymbol = "☢️ "

// A Querier queries and contains information about a Rust server
type Querier struct {
	ServerConfig
	tcpAddr    net.TCPAddr
	Name       string
	PlayerInfo PlayerInfo
}

// NewQuerier creats a new Server for observing a Rust server
func NewQuerier(server ServerConfig) (*Querier, error) {
	var sq = Querier{}
	rustIP, err := net.ResolveIPAddr("ip", server.Hostname)
	if err != nil {
		log.Println(qLogSymbol + err.Error())
		return nil, err
	}
	sq.tcpAddr = net.TCPAddr{IP: rustIP.IP, Port: server.Port}

	return &sq, nil
}

// Update queries a Rust server and updates Server with it's new information
func (sq *Querier) Update() error {
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
