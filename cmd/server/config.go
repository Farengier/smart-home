package main

import (
	"fmt"
	"time"
)

type YamlConfig struct {
	Log    LogConfig    `yaml:"log"`
	Server ServerConfig `yaml:"server"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type ServerConfig struct {
	Listen  string `yaml:"listen"`
	Port    int    `yaml:"port"`
	ReadTO  int    `yaml:"read_timeout"`
	WriteTO int    `yaml:"write_timeout"`
}

func (sc ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", sc.Listen, sc.Port)
}
func (sc ServerConfig) ReadTimeout() time.Duration {
	return time.Duration(sc.ReadTO) * time.Second
}
func (sc ServerConfig) WriteTimeout() time.Duration {
	return time.Duration(sc.WriteTO) * time.Second
}
