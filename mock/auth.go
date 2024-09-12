package mock

import (
	"context"

	"github.com/maliByatzes/ocs"
)

var _ ocs.AuthService = (*AuthService)(nil)

type AuthService struct {
	FindAuthByIDFn func(ctx context.Context, id int) (*ocs.Auth, error)
	FindAuthsFn    func(ctx context.Context, filter ocs.AuthFilter) ([]*ocs.Auth, int, error)
	CreateAuthFn   func(ctx context.Context, auth *ocs.Auth) error
	DeleteAuthFn   func(ctx context.Context, id int) error
}

func (a *AuthService) FindAuthByID(ctx context.Context, id int) (*ocs.Auth, error) {
	return a.FindAuthByID(ctx, id)
}

func (a *AuthService) FindAuths(ctx context.Context, filter ocs.AuthFilter) ([]*ocs.Auth, int, error) {
	return a.FindAuths(ctx, filter)
}

func (a *AuthService) CreateAuth(ctx context.Context, auth *ocs.Auth) error {
	return a.CreateAuth(ctx, auth)
}

func (a *AuthService) DeleteAuth(ctx context.Context, id int) error {
	return a.DeleteAuth(ctx, id)
}
