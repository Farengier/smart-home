package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Farengier/smart-home/internal/db"
	"github.com/Farengier/smart-home/internal/img"
	"github.com/Farengier/smart-home/internal/orm"
	"github.com/Farengier/smart-home/internal/signal"
	"github.com/jltorresm/otpgo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"image/png"
	"io"
	"os"
)

var conf *string

func init() {
	conf = flag.String("config", "config.yml", "config file path")
}

type DB interface {
	GORM() *gorm.DB
}

func main() {
	flag.Parse()

	cfg, err := initConfig()
	if err != nil {
		panic(err)
	}

	err = initLogging(cfg.Log)
	if err != nil {
		fmt.Printf("Error log init: %s\n", err)
		os.Exit(1)
	}

	signal.Init()

	//testDB(cfg)
	testTotp()
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
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

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

func testDB(cfg *YamlConfig) {
	dbc, err := db.New(cfg.DataBase)
	if err != nil {
		panic(err)
	}
	login := "adminasd"
	usr := &orm.User{}
	dbres := dbc.GORM().Joins("Role").First(usr, orm.User{Login: login})
	if !errors.Is(dbres.Error, gorm.ErrRecordNotFound) {
		fmt.Printf("Found\n")
		return
	} else {
		fmt.Printf("Not Found\n")
		return
	}
}

func testTotp() {
	totp := otpgo.TOTP{}
	_, err := totp.Generate()
	if err != nil {
		log.Errorf("[TG Bot Totp] key generate failed: %s", err)
		return
	}

	//key := totp.Key

	ku := totp.KeyUri("test", "test")
	qrcode, err := ku.QRCode()
	if err != nil {
		log.Errorf("[TG Bot Totp] key generate failed: %s", err)
		return
	}

	im, err := img.ParseB64(qrcode)
	if err != nil {
		log.Errorf("parse failed: %s", err)
		return
	}

	f, err := os.OpenFile("qr.png", os.O_WRONLY|os.O_CREATE, 0777)
	defer f.Close()
	if err != nil {
		log.Errorf("file open failed: %s", err)
		return
	}

	err = png.Encode(f, im)
	if err != nil {
		log.Errorf("file write failed: %s", err)
		return
	}
	//buf := bytes.NewBufferString("")
	//enc := base64.NewEncoder(base64.StdEncoding, buf)
	//fmt.Printf("full: %s\n", qrcode)
	//fmt.Printf("data: %s\n", qrcode[22:])
	//_, _ = enc.Write([]byte(qrcode[17:]))
	//os.WriteFile("qr.png", buf.Bytes(), 0777)
}

func clearDB(cfg *YamlConfig, dbc DB) {
	dbc.GORM().Migrator().DropTable(&orm.User{})
	dbc.GORM().Migrator().CreateTable(&orm.User{})
	dbc.GORM().Create(&orm.User{
		Model:  gorm.Model{ID: 1},
		Login:  "admin",
		OtpKey: cfg.AdminKey,
	})

	dbc.GORM().Migrator().DropTable(&orm.UserRole{})
	dbc.GORM().Migrator().CreateTable(&orm.UserRole{})
	dbc.GORM().Create(&orm.UserRole{
		Model:  gorm.Model{ID: 1},
		UserID: 1,
		Role:   "admin",
	})
}
