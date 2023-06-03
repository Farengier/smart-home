package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/Farengier/smart-home/internal/signal"
	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"time"
)

const syncCheckInterval = time.Second * 5
const syncMaxDuration = time.Minute
const syncChanBufferLen = 5

type Config interface {
	DBDirPath() string
	SyncInterval() time.Duration
	Backups() int
}

type db struct {
	cfg      Config
	dbDriver *sqlite3.SQLiteDriver
	dbc      *sql.DB
	gormDB   *gorm.DB
	ctx      context.Context

	t            *time.Ticker
	lastSyncTime time.Time
	syncCh       chan struct{}
}

func New(cfg Config) (*db, error) {
	d := &db{
		cfg:          cfg,
		lastSyncTime: time.Now(),
		dbDriver:     &sqlite3.SQLiteDriver{},
		t:            time.NewTicker(syncCheckInterval),
		syncCh:       make(chan struct{}, syncChanBufferLen),
	}

	ctx, cncl := context.WithCancel(context.Background())
	d.ctx = ctx
	signal.OnShutdown(func() error {
		cncl()
		d.t.Stop()
		return nil
	})

	dbc, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("db failed creating memory connection: %w", err)
	}
	d.dbc = dbc

	gormLog := logger.New(
		log.StandardLogger(),
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)
	d.gormDB, err = gorm.Open(gormSqlite.Dialector{Conn: d.dbc}, &gorm.Config{Logger: gormLog})
	if err != nil {
		return nil, fmt.Errorf("db failed gorm-ing connection: %w", err)
	}

	err = d.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}
	return d, nil
}

func (d *db) SyncNow() {
	d.syncCh <- struct{}{}
}
func (d *db) SqlDB() *sql.DB {
	return d.dbc
}

func (d *db) GORM() *gorm.DB {
	return d.gormDB
}

func (d *db) init(ctx context.Context) error {
	err := d.syncUp(ctx)
	if err != nil {
		return fmt.Errorf("sync up failed: %w", err)
	}

	signal.Run(d.syncer)
	return nil
}

func (d *db) syncUp(ctx context.Context) error {
	files, err := os.ReadDir(d.cfg.DBDirPath())
	if err != nil {
		return fmt.Errorf("reading dir failed: %w", err)
	}

	// choosing last database file with prefix and tipe (list is sorted)
	var file os.DirEntry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), "db_") || !strings.HasSuffix(f.Name(), ".sqlite") {
			continue
		}
		file = f
	}
	if file == nil {
		log.Infof("[DB] no stored database in ir %s", d.cfg.DBDirPath())
		return nil
	}

	log.Infof("[DB] starting sync up from %s", file.Name())

	dsn := fmt.Sprintf("file:%s/%s?mode=rw", d.cfg.DBDirPath(), file.Name())
	sdbc, err := d.dbDriver.Open(dsn)
	if err != nil {
		return fmt.Errorf("failed connecting to db %s: %w", dsn, err)
	}
	defer func(dbsc driver.Conn) {
		_ = dbsc.Close()
	}(sdbc)

	sdb, ok := sdbc.(*sqlite3.SQLiteConn)
	if !ok {
		return fmt.Errorf("failed assetring source connection as sqlite")
	}

	tdbc, err := d.dbc.Conn(ctx)
	if err != nil {
		return fmt.Errorf("memory connection failed")
	}
	defer func(tdbc *sql.Conn) {
		_ = tdbc.Close()
	}(tdbc)

	// backup
	err = tdbc.Raw(func(targetConn any) error {
		tdb, ok := targetConn.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("failed assetring memory connection as sqlite")
		}

		bck, err := tdb.Backup("main", sdb, "main")
		if err != nil {
			return fmt.Errorf("sync up backup start failed: %w", err)
		}

		log.Infof("[DB] sync up start")
		for {
			ok, err := bck.Step(1)
			if err != nil {
				log.Errorf("[DB] sync up error: %w", err)
			}
			if ok {
				log.Infof("[DB] sync up finishing")
				break
			}
		}
		err = bck.Finish()
		if err != nil {
			log.Errorf("[DB] sync done with error: %w", err)
			return nil
		}
		log.Infof("[DB] sync done")
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync up failed: %w", err)
	}
	return nil
}

func (d *db) syncer() {
	log.Info("[DB] running syncer")
	for {
		select {
		case <-d.ctx.Done():
			log.Info("[DB] sync down by closed context")
			err := d.syncDown()
			if err != nil {
				log.Errorf("[DB] sync down failed: %s", err)
			}
			return
		case <-d.syncCh:
			log.Info("[DB] sync down by channel")
			err := d.syncDown()
			if err != nil {
				log.Errorf("[DB] sync down failed: %s", err)
			}
		case <-d.t.C:
			if d.lastSyncTime.Sub(time.Now()) <= d.cfg.SyncInterval() {
				continue
			}
			log.Info("[DB] sync down by timer")
			err := d.syncDown()
			if err != nil {
				log.Errorf("[DB] sync down failed: %s", err)
			}
		}
	}
}

func (d *db) syncDown() error {
	log.Info("[DB] Vacuuming")
	_, err := d.dbc.Exec("VACUUM")
	if err != nil {
		log.Errorf("[DB] memory vacuum failed: %s", err)
	}

	dsn := fmt.Sprintf("file:%s/db_%s.sqlite?mode=rwc", d.cfg.DBDirPath(), time.Now().Format("2006_01_02_15_04_05"))
	dbtc, err := d.dbDriver.Open(dsn)
	if err != nil {
		return fmt.Errorf("failed connecting to db %s: %w", dsn, err)
	}
	defer func(dbtc driver.Conn) {
		_ = dbtc.Close()
	}(dbtc)

	tdb, ok := dbtc.(*sqlite3.SQLiteConn)
	if !ok {
		return fmt.Errorf("failed assetring source connection as sqlite")
	}

	ctx, cncl := context.WithTimeout(context.Background(), syncMaxDuration)
	defer cncl()

	sdbc, err := d.dbc.Conn(ctx)
	if err != nil {
		return fmt.Errorf("memory connection failed")
	}
	defer func(sdbc *sql.Conn) {
		_ = sdbc.Close()
	}(sdbc)

	// backup
	err = sdbc.Raw(func(sourceConn any) error {
		sdb, ok := sourceConn.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("failed assetring memory connection as sqlite")
		}

		bck, err := tdb.Backup("main", sdb, "main")
		if err != nil {
			return fmt.Errorf("sync down backup start failed: %w", err)
		}

		log.Infof("[DB] sync down start to %s", dsn)
		for {
			ok, err := bck.Step(1)
			if err != nil {
				log.Errorf("[DB] sync down error: %w", err)
			}
			if ok {
				log.Infof("[DB] sync down finishing")
				break
			}
		}
		err = bck.Finish()
		if err != nil {
			log.Errorf("[DB] sync done with error: %w", err)
			return nil
		}
		log.Infof("[DB] sync down done")
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync down failed: %w", err)
	}

	d.clearExtraDbs()
	return nil
}

func (d *db) clearExtraDbs() {
	files, err := os.ReadDir(d.cfg.DBDirPath())
	if err != nil {
		log.Errorf("[DB] remove old dbs failed: %s", err)
	}

	var dbs []os.DirEntry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), "db_") || !strings.HasSuffix(f.Name(), ".sqlite") {
			continue
		}
		dbs = append(dbs, f)
	}
	if len(dbs) > d.cfg.Backups() {
		toClear := dbs[0 : len(dbs)-d.cfg.Backups()]
		for _, f := range toClear {
			fn := fmt.Sprintf("%s/%s", d.cfg.DBDirPath(), f.Name())
			log.Infof("[DB] removing old db %s", fn)
			err = os.Remove(fn)
			if err != nil {
				log.Errorf("[DB] failed removing %s: %s", fn, err)
			}
		}
	}
}
