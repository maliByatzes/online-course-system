package ocs

import "time"

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
