package sqlite_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/maliByatzes/ocs/sqlite"
)

var dump = flag.Bool("dump", false, "save work data")

func TestDB(t *testing.T) {
	db := MustOpenDB(t)
	MustCloseDB(t, db)
}

func MustOpenDB(tb testing.TB) *sqlite.DB {
	tb.Helper()

	dsn := ":memory:"
	if *dump {
		dir, err := os.MkdirTemp("", "")
		if err != nil {
			tb.Fatal(err)
		}
		dsn = filepath.Join(dir, "db")
		println("DUMP=" + dsn)
	}

	db := sqlite.NewDB(dsn)
	if err := db.Open(); err != nil {
		tb.Fatal(err)
	}
	return db
}

func MustCloseDB(tb testing.TB, db *sqlite.DB) {
	tb.Helper()
	if err := db.Close(); err != nil {
		tb.Fatal(err)
	}
}
