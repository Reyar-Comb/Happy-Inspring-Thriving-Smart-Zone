package main

import (
	"net"

	"github.com/Reyar-Comb/HITPlane/config"
	"github.com/Reyar-Comb/HITPlane/server"
)

func main() {
	config.InitConfig()

	room := server.NewRoom()

	s := &server.Server{
		Addr:  net.JoinHostPort("", config.GlobalConfig.Port),
		Rooms: []*server.Room{room},
	}

	err := server.Start(s)
	if err != nil {
		panic(err)
	}
}
