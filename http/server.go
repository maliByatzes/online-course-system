package http

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/maliByatzes/ocs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocs_http_request_count",
		Help: "Total number of requests by route",
	}, []string{"method", "path"})

	requestSeconds = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocs_http_request_seconds",
		Help: "Total amount of requests time by route, in seconds",
	}, []string{"method", "path"})
)

const ShutdownTimeout = 1 * time.Second

type Server struct {
	ln                 net.Listener
	server             *http.Server
	router             *mux.Router
	sc                 *securecookie.SecureCookie
	Addr               string
	Domain             string
	HashKey            string
	BlockKey           string
	GithubClientID     string
	GithubClientSecret string
	AuthService        ocs.AuthService
	StudentService     ocs.StudentService
}
