package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	UDPPort  string `yaml:"udp_port" env-default:":8888"`
	HTTPPort string `yaml:"http_port" env-default:":8889"`
	Timeout  string `yaml:"timeout" env-default:"5s"`
}

var GlobalConfig *Config

func InitConfig() {
	instance := &Config{}
	err := cleanenv.ReadConfig("config.yaml", instance)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	GlobalConfig = instance
	fmt.Printf("Config: %+v\n", GlobalConfig)
}
