package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	studentCountGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ocs_db_students",
		Help: "The total number of students",
	})
)

var migrationFS embed.FS

type DB struct {
	db           *sql.DB
	ctx          context.Context
	cancel       func()
	DSN          string
	EventService ocs.EventService
	Now          func() time.Time
}
