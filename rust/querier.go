package rust

import (
	"log"
	"net"
	"time"

	"github.com/alliedmodders/blaster/valve"
)

const qLogSymbol = "☢️ "

type QueryProvider interface {
	QueryInfo() (*valve.ServerInfo, error)
	Close()
}

type QueryInfo struct {
	Name       string
	Players    uint8
	MaxPlayers uint8
}

// A Querier queries and contains information about a Rust server
type Querier struct {
	ServerConfig
	tcpAddr    net.TCPAddr
	Name       string
	PlayerInfo PlayerInfo
	provider   QueryProvider
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

	query, err := valve.NewServerQuerier(sq.tcpAddr.String(), time.Second*3)
	if err != nil {
		return nil, err
	}

	sq.provider = query

	return &sq, nil
}

// Update queries a Rust server and updates Server with it's new information
func (sq *Querier) Update() error {
	info, err := sq.provider.QueryInfo()
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

func (sq *Querier) Close() {
	if sq.provider != nil {
		sq.provider.Close()
	}
}
