package http

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/maliByatzes/ocs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
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
	EventService       ocs.EventService
	StudentService     ocs.StudentService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	s.router.Use(reportPanic)
	s.server.Handler = http.HandlerFunc(s.serveHTTP)
	s.router.NotFoundHandler = http.HandlerFunc(s.handleNotFound)

	s.router.HandleFunc("/debug/version", s.handleVersion).Methods("GET")
	s.router.HandleFunc("/debug/commit", s.handleCommit).Methods("GET")

	router := s.router.PathPrefix("/").Subrouter()
	router.Use(s.authenticate)

	// Non-auth routes
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(s.requireNoAuth)
		// s.registerAuthRoutes(r)
	}

	// Auth routes
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(s.requireAuth)
		//s.registerEventRoutes(r)
	}

	return s
}

func (s *Server) UseTLS() bool {
	return s.Domain != ""
}

func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}
	return "http"
}

func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	// localhost
	domain := "127.0.0.1"
	if s.Domain != "" {
		domain = s.Domain
	}

	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", s.Scheme(), domain)
	}
	return fmt.Sprintf("%s://%s:%d", s.Scheme(), domain, s.Port())
}

func (s *Server) Open() (err error) {
	if err := s.openSecureCookie(); err != nil {
		return err
	}

	if s.GithubClientID == "" {
		return fmt.Errorf("github client id required")
	} else if s.GithubClientSecret == "" {
		return fmt.Errorf("github client secret required")
	}

	if s.Domain != "" {
		s.ln = autocert.NewListener(s.Domain)
	} else {
		if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
			return err
		}
	}

	// Serve isntead of ListenAndServe
	go s.server.Serve(s.ln)

	return nil
}

func (s *Server) openSecureCookie() error {
	if s.HashKey == "" {
		return fmt.Errorf("hash key required")
	} else if s.BlockKey == "" {
		return fmt.Errorf("block key required")
	}

	hashKey, err := hex.DecodeString(s.HashKey)
	if err != nil {
		return fmt.Errorf("invalid hash key")
	}

	blockKey, err := hex.DecodeString(s.BlockKey)
	if err != nil {
		return fmt.Errorf("invalid block key")
	}

	s.sc = securecookie.New(hashKey, blockKey)
	s.sc.SetSerializer(securecookie.JSONEncoder{})

	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.GithubClientID,
		ClientSecret: s.GithubClientSecret,
		Scopes:       []string{},
		Endpoint:     github.Endpoint,
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		switch v := r.PostFormValue("_method"); v {
		case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
			r.Method = v
		}
	}

	switch ext := path.Ext(r.URL.Path); ext {
	case ".json":
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")
		r.URL.Path = strings.TrimSuffix(r.URL.Path, ext)
	case ".csv":
		r.Header.Set("Accept", "text/csv")
		r.URL.Path = strings.TrimSuffix(r.URL.Path, ext)
	}

	s.router.ServeHTTP(w, r)
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("Authorization"); strings.HasPrefix(v, "Bearer ") {
			apiKey := strings.TrimPrefix(v, "Bearer ")

			students, _, err := s.StudentService.FindStudents(r.Context(), ocs.StudentFilter{APIKey: &apiKey})
			if err != nil {
				Error(w, r, err)
				return
			} else if len(students) == 0 {
				Error(w, r, ocs.Errorf(ocs.EUNAUTHORIZED, "Invalid API key"))
				return
			}

			r = r.WithContext(ocs.NewContextWithStudent(r.Context(), students[0]))

			next.ServeHTTP(w, r)
			return
		}

		session, _ := s.session(r)

		if session.StudentID != 0 {
			if student, err := s.StudentService.FindStudentByID(r.Context(), session.StudentID); err != nil {
				log.Printf("cannot find session student: id=%d err=%s", session.StudentID, err)
			} else {
				r = r.WithContext(ocs.NewContextWithStudent(r.Context(), student))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireNoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if studentID := ocs.StudentIDFromContext(r.Context()); studentID != 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusFound)
			data, err := json.Marshal(struct {
				message string
			}{
				message: "Student is already logged in.",
			})
			if err != nil {
				log.Printf("http: cannot marshal data: %s", err)
			}
			w.Write(data)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if studentID := ocs.StudentIDFromContext(r.Context()); studentID != 0 {
			next.ServeHTTP(w, r)
			return
		}

		redirectURL := r.URL
		redirectURL.Scheme, redirectURL.Host = "", ""

		session, _ := s.session(r)
		session.RedirectURL = redirectURL.String()
		if err := s.setSession(w, session); err != nil {
			log.Printf("http: cannot set session: %s", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusFound)
		data, err := json.Marshal(struct {
			error string
		}{
			error: "You are not logged in",
		})
		if err != nil {
			log.Printf("http: cannot marshal json: %s", err)
		}
		w.Write(data)
		// http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func loadFlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, _ := r.Cookie("flash"); cookie != nil {
			SetFlash(w, "")
			r = r.WithContext(ocs.NewContextWithFlash(r.Context(), cookie.Value))
		}

		next.ServeHTTP(w, r)
	})
}

func reportPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				ocs.ReportPanic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Your page is not found")
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ocs.Version))
}

func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ocs.Commit))
}

func (s *Server) session(r *http.Request) (Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return Session{}, err
	}

	var session Session
	if err := s.UnmarshalSession(cookie.Value, &session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *Server) setSession(w http.ResponseWriter, session Session) error {
	buf, err := s.MarshalSession(session)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    buf,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Secure:   s.UseTLS(),
		HttpOnly: true,
	})
	return nil
}

func (s *Server) MarshalSession(session Session) (string, error) {
	return s.sc.Encode(SessionCookieName, session)
}

func (s *Server) UnmarshalSession(data string, session *Session) error {
	return s.sc.Decode(SessionCookieName, data, &session)
}

func ListenAndServeTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}

func ListenAndServeDebug() error {
	h := http.NewServeMux()
	h.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":6060", h)
}
