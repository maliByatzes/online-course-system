package http

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/maliByatzes/ocs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	errorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ocs_http_error_count",
		Help: "Total Number of errors by error code",
	}, []string{"code"})
)

type Client struct {
	URL string
}

func NewClient(u string) *Client {
	return &Client{URL: u}
}

func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.URL+url, body)
	if err != nil {
		return nil, err
	}

	if student := ocs.StudentFromContext(ctx); student != nil && student.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+student.APIKey)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

const sessionCookieName = "cookie"

type Session struct {
	studentID   int    `json:"studentID"`
	RedirectURL string `json:"redirectURL"`
	State       string `json:"state"`
}

func SetFlash(w http.ResponseWriter, s string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    s,
		Path:     "/",
		HttpOnly: true,
	})
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	code, message := ocs.ErrorCode(err), ocs.ErrorMessage(err)
	errorCount.WithLabelValues(code).Inc()

	if code == ocs.EINTERNAL {
		ocs.ReportError(r.Context(), err, r)
		LogError(r, err)
	}

	switch r.Header.Get("Accept") {
	case "application/json":
		w.Header().Set("Accept", "application/json")
		w.WriteHeader(ErrorStatusCode(code))
		json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
	default:
		// Encode all responses as json for now
		w.Header().Set("Accept", "application/json")
		w.WriteHeader(ErrorStatusCode(code))
		json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func parseErrorResponse(resp *http.Response) error {
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var errorResponse ErrorResponse
	if err := json.Unmarshal(buf, &errorResponse); err != nil {
		message := strings.TrimSpace(string(buf))
		if message == "" {
			message = "Empty response from server."
		}
		return ocs.Errorf(FromErrorStatusCode(resp.StatusCode), message)
	}
	return ocs.Errorf(FromErrorStatusCode(resp.StatusCode), errorResponse.Error)
}

func LogError(r *http.Request, err error) {
	log.Printf("[http] error: %s %s: %s", r.Method, r.URL, err)
}

var codes = map[string]int{
	ocs.ECONFLICT:       http.StatusConflict,
	ocs.EINVALID:        http.StatusBadRequest,
	ocs.ENOTFOUND:       http.StatusNotFound,
	ocs.ENOTIMPLEMENTED: http.StatusNotImplemented,
	ocs.EUNAUTHORIZED:   http.StatusUnauthorized,
	ocs.EINTERNAL:       http.StatusInternalServerError,
}

func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return ocs.EINTERNAL
}
