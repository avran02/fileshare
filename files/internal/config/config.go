package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Minio  Minio  `yaml:"minio"`
	Server Server `yaml:"server"`
}

type Minio struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}

type Server struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
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
