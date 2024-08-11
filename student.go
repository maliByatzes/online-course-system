package ocs

import (
	"context"
	"time"
)

type Student struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	APIKey    string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Auths     []*Auth   `json:"auths"`
}

func (s *Student) Validate() error {
	if s.Name == "" || s.Email == "" {
		return Errorf(EINVALID, "Provide required fields")
	}
	return nil
}

func (s *Student) AvatarURL(size int) string {
	for _, auth := range s.Auths {
		if s := auth.AvatarURL(size); s != "" {
			return s
		}
	}
	return ""
}

type StudentService interface {
	FindStudentByID(ctx context.Context, id int) (*Student, error)
	FindStudents(ctx context.Context, filter StudentFilter) ([]*Student, error)
	CreateStudent(ctx context.Context, student *Student) error
	UpdateStudent(ctx context.Context, id int, upd StudentUpdate) (*Student, error)
	DeleteStudent(ctx context.Context, id int) error
}

type StudentFilter struct {
	ID     *int    `json:"id"`
	Email  *string `json:"email"`
	APIKey *string `json:"apiKey"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}

type StudentUpdate struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}
