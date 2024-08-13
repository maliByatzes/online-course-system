package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	auth, err := findAuthByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachAuthAssociations(ctx, tx, auth); err != nil {
		return nil, err
	}

	return auth, nil
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

func findAuthByID(ctx context.Context, tx *Tx, id int) (*ocs.Auth, error) {
	auths, _, err := findAuths(ctx, tx, ocs.AuthFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(auths) == 0 {
		return nil, &ocs.Error{Code: ocs.ENOTFOUND, Message: "Auth not found."}
	}
	return auths[0], nil
}

func findAuths(ctx context.Context, tx *Tx, filter ocs.AuthFilter) (_ []*ocs.Auth, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.StudentID; v != nil {
		where, args = append(where, "student_id = ?"), append(args, *v)
	}
	if v := filter.Source; v != nil {
		where, args = append(where, "source = ?"), append(args, *v)
	}
	if v := filter.SourceID; v != nil {
		where, args = append(where, "source_id = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
    SELECT
      id,
      student_id,
      source,
      source_id,
      access_token,
      refresh_token,
      expiry,
      created_at,
      updated_at,
      COUNT(*) OVER()
    FROM auths
    WHERE `+strings.Join(where, " AND ")+`
    ORDER BY id ASC
    `+FormatLimitOffset(filter.Limit, filter.Offset)+`
    `,
		args...,
	)
	if err != nil {
		return nil, n, FormatError(err)
	}
	defer rows.Close()

	auths := make([]*ocs.Auth, 0)
	for rows.Next() {
		var auth ocs.Auth
		var expiry sql.NullString
		if err := rows.Scan(
			&auth.ID,
			&auth.StudentID,
			&auth.Source,
			&auth.SourceID,
			&auth.AccessToken,
			&auth.RefreshToken,
			&expiry,
			(*NullTime)(&auth.CreatedAt),
			(*NullTime)(&auth.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}

		if expiry.Valid {
			if v, _ := time.Parse(time.RFC3339, expiry.String); !v.IsZero() {
				auth.Expiry = &v
			}
		}

		auths = append(auths, &auth)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, FormatError(err)
	}

	return auths, n, nil
}

func attachAuthAssociations(ctx context.Context, tx *Tx, auth *ocs.Auth) (err error) {
	if auth.Student, err = findStudentByID(ctx, tx, auth.StudentID); err != nil {
		return fmt.Errorf("attach auth user: %w", err)
	}
	return nil
}
