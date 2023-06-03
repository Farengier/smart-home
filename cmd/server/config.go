package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type YamlConfig struct {
	Log      LogConfig    `yaml:"log"`
	Server   ServerConfig `yaml:"server"`
	Telegram TBotConfig   `yaml:"telegram"`
	DataBase DBConfig     `yaml:"db"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type TBotConfig struct {
	BotToken string `yaml:"token"`
	TimeOut  struct {
		Sensitive string `yaml:"sensitive"`
		Low       string `yaml:"low"`
	} `yaml:"spam_timeout"`
}

func (tbc TBotConfig) Token() string {
	return tbc.BotToken
}
func (tbc TBotConfig) SpamFilterDurationSensitive() time.Duration {
	d, err := time.ParseDuration(tbc.TimeOut.Sensitive)
	if err != nil {
		log.Errorf("[Config] wrong tg spam filter sensitive interval format: %s", err)
		d = time.Second * 60
	}
	return d
}
func (tbc TBotConfig) SpamFilterDurationLow() time.Duration {
	d, err := time.ParseDuration(tbc.TimeOut.Low)
	if err != nil {
		log.Errorf("[Config] wrong tg spam filter low interval format: %s", err)
		d = time.Second * 1
	}
	return d
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

type DBConfig struct {
	Path      string `yaml:"path"`
	BackupCnt int    `yaml:"backups"`
	Sync      string `yaml:"sync"`
	syncDur   time.Duration
}

func (dbc DBConfig) DBDirPath() string {
	return dbc.Path
}
func (dbc DBConfig) SyncInterval() time.Duration {
	if dbc.syncDur != 0 {
		return dbc.syncDur
	}

	d, err := time.ParseDuration(dbc.Sync)
	if err != nil {
		log.Errorf("[Config] wrong db sync interval format: %s", err)
		d = time.Hour * 24
	}
	dbc.syncDur = d

	return dbc.syncDur
}
func (dbc DBConfig) Backups() int {
	return dbc.BackupCnt
}
