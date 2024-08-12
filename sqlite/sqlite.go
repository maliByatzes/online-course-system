package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/maliByatzes/ocs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	studentCountGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ocs_db_students",
		Help: "The total number of students",
	})
)

//go:embed migration/*.sql
var migrationFS embed.FS

type DB struct {
	db           *sql.DB
	ctx          context.Context
	cancel       func()
	DSN          string
	EventService ocs.EventService
	Now          func() time.Time
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN:          dsn,
		Now:          time.Now,
		EventService: ocs.NopEventService(),
	}

	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	if db.DSN != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(db.DSN), 0700); err != nil {
			return err
		}
	}

	if db.db, err = sql.Open("sqlite3", db.DSN); err != nil {
		return err
	}

	if _, err := db.db.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		return fmt.Errorf("enable wal: %w", err)
	}

	if _, err := db.db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("foreign keys pragma: %w", err)
	}

	if err := db.migrate(); err != nil {
		return err
	}

	go db.monitor()

	return nil
}

func (db *DB) migrate() error {
	return nil
}

func (db *DB) monitor() {}
