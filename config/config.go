package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port    string `yaml:"port" env-default:":8888"`
	Timeout string `yaml:"timeout" env-default:"5s"`
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
