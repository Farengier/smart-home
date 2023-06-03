package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type YamlConfig struct {
	AdminKey string    `yaml:"adminKey"`
	Log      LogConfig `yaml:"log"`
	DataBase DBConfig  `yaml:"db"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
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
