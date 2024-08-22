package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	websocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ocs_http_websocket_connections",
		Help: "Total number of connected websocket users",
	})
)

func (s *Server) registerEventRoutes(r *mux.Router) {
	r.HandleFunc("/events", s.handleEvents)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	websocketConnections.Inc()
	defer websocketConnections.Dec()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		LogError(r, err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	r = r.WithContext(ctx)
	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})

	defer conn.Close()

	go ignoreWebSocketReaders(conn)

	sub, err := s.EventService.Subscribe(r.Context())
	if err != nil {
		LogError(r, err)
		return
	}
	defer sub.Close()

	for {
		select {
		case <-r.Context().Done():
			return

		case event, ok := <-sub.C():
			if !ok {
				return
			}

			buf, err := json.Marshal(event)
			if err != nil {
				LogError(r, err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, buf); err != nil {
				LogError(r, err)
				return
			}
		}
	}
}

func ignoreWebSocketReaders(conn *websocket.Conn) {
	for {
		if _, _, err := conn.NextReader(); err != nil {
			conn.Close()
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
