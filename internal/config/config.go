package config

import "os"

type API interface {
	Get(key string) string
}

func NewConfig() *Config {
	return &Config{}
}

type Config struct {
}

func (c *Config) Get(key string) string {
	return os.Getenv(key)
}
