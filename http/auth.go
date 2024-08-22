package http

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v63/github"
	"github.com/gorilla/mux"
	"github.com/maliByatzes/ocs"
	"golang.org/x/oauth2"
)

func (s *Server) registerAuthRoutes(r *mux.Router) {
	// r.HandleFunc("/login", s.handleLogin).Methods("GET")
	r.HandleFunc("/logout", s.handleLogout).Methods("DELETE")
	r.HandleFunc("/oauth/github", s.handleOAuthGithub).Methods("GET")
	r.HandleFunc("/oauth/github/callback", s.handleOAuthGithubCallback).Methods("GET")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := s.setSession(w, Session{}); err != nil {
		Error(w, r, err)
		return
	}

	sendJsonResponseMessage(w, "Logged out successfully.")
}

func (s *Server) handleOAuthGithub(w http.ResponseWriter, r *http.Request) {
	session, err := s.session(r)
	if err != nil {
		Error(w, r, err)
		return
	}

	state := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, state); err != nil {
		Error(w, r, err)
		return
	}
	session.State = hex.EncodeToString(state)

	if err := s.setSession(w, session); err != nil {
		Error(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusFound)
	data, err := json.Marshal(struct {
		message string
		url     string
	}{
		message: "Access OAuth2 provider link.",
		url:     s.OAuth2Config().AuthCodeURL(session.State),
	})
	if err != nil {
		log.Printf("http: cannot marshal data: %s", err)
	}
	w.Write(data)
}

func (s *Server) handleOAuthGithubCallback(w http.ResponseWriter, r *http.Request) {
	state, code := r.FormValue("state"), r.FormValue("code")

	session, err := s.session(r)
	if err != nil {
		Error(w, r, fmt.Errorf("cannot read session: %s", err))
		return
	}

	if state != session.State {
		Error(w, r, fmt.Errorf("oauth sate mismatch"))
		return
	}

	tok, err := s.OAuth2Config().Exchange(r.Context(), code)
	if err != nil {
		Error(w, r, fmt.Errorf("oauth exachange error: %s", err))
		return
	}

	client := github.NewClient(oauth2.NewClient(r.Context(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok.AccessToken},
	)))

	u, _, err := client.Users.Get(r.Context(), "")
	if err != nil {
		Error(w, r, fmt.Errorf("cannot fetch github user: %s", err))
		return
	} else if u.ID == nil {
		Error(w, r, fmt.Errorf("user ID not returned by Github, cannot authenticate user"))
		return
	}

	var name string
	if u.Name != nil {
		name = *u.Name
	} else if u.Login != nil {
		name = *u.Login
	}

	// WARNING: Email is always required by the db
	var email string
	if u.Email != nil {
		email = *u.Email
	}

	auth := &ocs.Auth{
		Source:       ocs.AuthSourceGithub,
		SourceID:     strconv.FormatInt(*u.ID, 10),
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Student: &ocs.Student{
			Name:  name,
			Email: email,
		},
	}
	if !tok.Expiry.IsZero() {
		auth.Expiry = &tok.Expiry
	}

	if err := s.AuthService.CreateAuth(r.Context(), auth); err != nil {
		Error(w, r, fmt.Errorf("cannot create auth: %s", err))
		return
	}

	redirectURL := session.RedirectURL

	session.StudentID = auth.ID
	session.RedirectURL = ""
	session.State = ""
	if err := s.setSession(w, session); err != nil {
		Error(w, r, fmt.Errorf("cannot set session cookie: %s", err))
		return
	}

	if redirectURL == "" {
		redirectURL = "/"
	}

	w.WriteHeader(http.StatusFound)
	sendJsonResponseMessage(w, "Logged in successfully")
}
