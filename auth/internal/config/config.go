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

type DB struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type JWT struct {
	Secret string `yaml:"secret"`
	Exp    int    `yaml:"exp"`
}

type Config struct {
	Debug  bool   `yaml:"debug"`
	Server Server `yaml:"server"`
	DB     DB     `yaml:"db"`
	JWT    JWT    `yaml:"jwt"`
}

func New() *Config {
	confFile, err := os.Open("config.yml")
	if err != nil {
		log.Fatal("can't open config file:\n", err)
	}
	defer confFile.Close()

	decoder := yaml.NewDecoder(confFile)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		confFile.Close()
		log.Fatal("can't decode config file:\n", err) //nolint
	}

	return config
}
