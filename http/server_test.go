package http_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/maliByatzes/ocs"
	ocshttp "github.com/maliByatzes/ocs/http"
	"github.com/maliByatzes/ocs/mock"
)

const (
	TestHashKey            = "0000000000000000"
	TestBlockKey           = "00000000000000000000000000000000"
	TestGithunClientID     = "00000000000000000000"
	TestGithubClientSecret = "0000000000000000000000000000000000000000"
)

type Server struct {
	*ocshttp.Server

	// Mock services
	AuthService    mock.AuthService
	EventService   mock.EventService
	StudentService mock.StudentService
}

func MustOpenServer(tb testing.TB) *Server {
	tb.Helper()

	s := &Server{Server: ocshttp.NewServer()}
	s.HashKey = TestHashKey
	s.BlockKey = TestBlockKey
	s.GithubClientID = TestGithunClientID
	s.GithubClientSecret = TestGithubClientSecret

	s.Server.AuthService = &s.AuthService
	s.Server.EventService = &s.EventService
	s.Server.StudentService = &s.StudentService

	if err := s.Open(); err != nil {
		tb.Fatal(err)
	}
	return s
}

func MustCloseServer(tb testing.TB, s *Server) {
	tb.Helper()
	if err := s.Close(); err != nil {
		tb.Fatal(err)
	}
}

func (s *Server) MustNewRequest(tb testing.TB, ctx context.Context, method, url string, body io.Reader) *http.Request {
	tb.Helper()

	r, err := http.NewRequest(method, s.URL()+url, body)
	if err != nil {
		tb.Fatal(err)
	}

	if student := ocs.StudentFromContext(ctx); student != nil {
		data, err := s.MarshalSession(ocshttp.Session{StudentID: student.ID})
		if err != nil {
			tb.Fatal(err)
		}
		r.AddCookie(&http.Cookie{
			Name:  ocshttp.SessionCookieName,
			Value: data,
			Path:  "/",
		})
	}

	return r
}
