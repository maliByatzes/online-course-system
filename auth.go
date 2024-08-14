package ocs

import (
	"context"
	"fmt"
	"time"
)

const (
	AuthSourceGithub = "github"
)

type Auth struct {
	ID           int        `json:"id"`
	StudentID    int        `json:"studentID"`
	Student      *Student   `json:"student"`
	Source       string     `json:"source"`
	SourceID     string     `json:"sourceID"`
	AccessToken  string     `json:"-"`
	RefreshToken string     `json:"-"`
	Expiry       *time.Time `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (a *Auth) Error() error {
	if a.StudentID == 0 {
		return Errorf(EINVALID, "Student required.")
	} else if a.Source == "" {
		return Errorf(EINVALID, "Source required.")
	} else if a.SourceID == "" {
		return Errorf(EINVALID, "Source ID required.")
	} else if a.AccessToken == "" {
		return Errorf(EINVALID, "Access Token required.")
	}
	return nil
}

func (a *Auth) AvatarURL(size int) string {
	switch a.Source {
	case AuthSourceGithub:
		return fmt.Sprintf("https://avatars1.githubusercontent.com/u/%s?s=%d", a.SourceID, size)
	default:
		return ""
	}
}

type AuthService interface {
	FindAuthByID(ctx context.Context, id int) (*Auth, error)
	FindAuths(ctx context.Context, filter AuthFilter) ([]*Auth, int, error)
	CreateAuth(ctx context.Context, auth *Auth) error
	DeleteAuth(ctx context.Context, id int) error
}

type AuthFilter struct {
	ID        *int    `json:"id"`
	StudentID *int    `json:"studentID"`
	Source    *string `json:"source"`
	SourceID  *string `json:"sourceID"`
	Offset    int     `json:"offset"`
	Limit     int     `json:"limit"`
}
