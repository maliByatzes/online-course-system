package sqlite

import (
	"context"

	"github.com/maliByatzes/ocs"
)

var _ ocs.AuthService = (*AuthService)(nil)

type AuthService struct {
	db *DB
}

func NewAuthService(db *DB) *AuthService {
	return &AuthService{db: db}
}

func (a *AuthService) FindAuthByID(ctx context.Context, id int) (*ocs.Auth, error) {
	return nil, nil
}

func (a *AuthService) FindAuths(ctx context.Context, filter ocs.AuthFilter) ([]*ocs.Auth, error) {
	return nil, nil
}

func (a *AuthService) CreateAuth(ctx context.Context, auth *ocs.Auth) error {
	return nil
}

func (a *AuthService) DeleteAuth(ctx context.Context, id int) error {
	return nil
}
