package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Farengier/smart-home/internal/signal"
	"github.com/Farengier/smart-home/internal/web"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var conf *string

func init() {
	conf = flag.String("config", "config.yml", "config file path")
}

func main() {
	flag.Parse()

	cfg, err := initConfig()
	if err != nil {
		fmt.Printf("Error reading config: %s\n", err)
		fmt.Println("Usage server --config=<file_path>")
		fmt.Println()
		os.Exit(1)
	}

	err = initLogging(cfg.Log)
	if err != nil {
		fmt.Printf("Error log init: %s\n", err)
		os.Exit(1)
	}

	signal.Init()
	signal.Run(func() { web.Start(cfg.Server) })
	signal.Wait()
	log.Info("[Server] Closing")
}

func initConfig() (*YamlConfig, error) {
	if conf == nil || *conf == "" {
		return nil, fmt.Errorf("config param is empty")
	}

	fmt.Printf("config is %s\n", *conf)

	f, err := os.Open(*conf)
	if err != nil {
		return nil, fmt.Errorf("open failed: %w", err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	cfg := &YamlConfig{}
	err = dec.Decode(cfg)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}
	return cfg, nil
}

func initLogging(cfg LogConfig) error {
	var w io.Writer
	w = os.Stdout
	if cfg.Path != "" {
		f, err := os.Create(cfg.Path)
		if err != nil {
			return fmt.Errorf("creating log file failed: %w", err)
		}
		w = io.MultiWriter(w, f)
	}

	lvl, err := log.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("level parse failed: %w", err)
	}

	log.SetOutput(w)
	log.SetLevel(lvl)
	return nil
}
