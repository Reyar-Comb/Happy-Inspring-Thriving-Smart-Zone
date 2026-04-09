package main

import (
	"net"

	"github.com/Reyar-Comb/HITPlane/config"
	"github.com/Reyar-Comb/HITPlane/server"
)

func main() {
	config.InitConfig()

	s := &server.Server{
		Addr:           net.JoinHostPort("", config.GlobalConfig.Port),
		Rooms:          map[int32]*server.Room{},
		AvailableRooms: map[int32]*server.Room{},
	}

	err := server.Start(s)
	if err != nil {
		panic(err)
	}
}
