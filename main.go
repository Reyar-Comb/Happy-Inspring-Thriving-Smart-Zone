package main

import (
	"github.com/Reyar-Comb/HITPlane/config"
	"github.com/Reyar-Comb/HITPlane/server"
)

func main() {
	config.InitConfig()

	s := server.NewServer()

	go s.StartHTTP()

	err := s.StartUDP()
	if err != nil {
		panic(err)
	}
}
