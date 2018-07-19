package rust

import (
	"log"
	"net"
	"time"

	"github.com/alliedmodders/blaster/valve"
)

const logSymbol = "👁️‍🗨️ "

type Server struct {
	Hostname string
	Port     int
}

type PlayerInfo struct {
	Players      uint8
	MaxPlayers   uint8
	PlayersDelta int8
}

type ServerInfo struct {
	Server
	tcpAddr    net.TCPAddr
	Name       string
	PlayerInfo PlayerInfo
}

func NewServerInfo(server Server) (*ServerInfo, error) {
	var sq = ServerInfo{}
	rustIP, err := net.ResolveIPAddr("ip", server.Hostname)
	if err != nil {
		log.Println(logSymbol+"Could not resolve rust.alittlemercy.com", err)
		return nil, err
	}
	sq.tcpAddr = net.TCPAddr{IP: rustIP.IP, Port: server.Port}

	return &sq, nil
}

func (sq *ServerInfo) Update() error {
	query, err := valve.NewServerQuerier(sq.tcpAddr.String(), time.Second*3)
	if err != nil {
		log.Println(logSymbol + "Could not contact Rust server")
		sq.PlayerInfo = PlayerInfo{}
		return err
	}
	defer query.Close()

	info, err := query.QueryInfo()
	if err != nil {
		sq.PlayerInfo = PlayerInfo{}
		log.Printf(logSymbol+"Error connecting to Rust server: %s, %s", sq.tcpAddr.String(), err)
		return err
	}

	sq.Name = info.Name
	sq.PlayerInfo.PlayersDelta = (int8)(info.Players - sq.PlayerInfo.Players)
	sq.PlayerInfo.MaxPlayers = info.MaxPlayers
	sq.PlayerInfo.Players = info.Players
	return nil
}
