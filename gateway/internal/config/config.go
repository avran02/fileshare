package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type FileService struct {
	Endpoint string `yaml:"endpoint"`
}

type Config struct {
	Server      Server      `yaml:"server"`
	FileService FileService `yaml:"fileService"`
}

func New() *Config {
	confFile, err := os.Open("config.yml")
	if err != nil {
		log.Fatal("can't open config file:\n", err)
	}

	decoder := yaml.NewDecoder(confFile)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		confFile.Close()
		log.Fatal("can't decode config file:\n", err)
	}

	confFile.Close()
	return config
}
