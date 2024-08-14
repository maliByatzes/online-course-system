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

func (a *AuthService) FindAuths(ctx context.Context, filter ocs.AuthFilter) ([]*ocs.Auth, int, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	auths, n, err := findAuths(ctx, tx, filter)
	if err != nil {
		return auths, n, err
	}

	for _, auth := range auths {
		if err := attachAuthAssociations(ctx, tx, auth); err != nil {
			return auths, n, err
		}
	}
	return auths, n, nil
}

func (a *AuthService) CreateAuth(ctx context.Context, auth *ocs.Auth) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if other, err := findAuthBySourceID(ctx, tx, auth.Source, auth.SourceID); err == nil {
		if other, err := updateAuth(ctx, tx, other.ID, auth.AccessToken, auth.RefreshToken, auth.Expiry); err != nil {
			return fmt.Errorf("cannot update auth: id=%d err=%w", other.ID, err)
		} else if err := attachAuthAssociations(ctx, tx, other); err != nil {
			return err
		}

		*auth = *other
		return tx.Commit()
	} else if ocs.ErrorCode(err) != ocs.ENOTFOUND {
		return fmt.Errorf("cannot find auth by source student: %w", err)
	}

	if auth.StudentID == 0 && auth.Student != nil {
		if student, err := findStudentByEmail(ctx, tx, auth.Student.Email); err == nil {
			auth.Student = student
		} else if ocs.ErrorCode(err) == ocs.ENOTFOUND {
			if err := createStudent(ctx, tx, auth.Student); err != nil {
				return fmt.Errorf("cannot create student: %w", err)
			}
		} else {
			return fmt.Errorf("cannot find student by email: %w", err)
		}

		auth.StudentID = auth.Student.ID
	}

	if err := createAuth(ctx, tx, auth); err != nil {
		return err
	} else if err := attachAuthAssociations(ctx, tx, auth); err != nil {
		return err
	}

	return tx.Commit()
}

func (a *AuthService) DeleteAuth(ctx context.Context, id int) error {
	return nil
}

func createAuth(ctx context.Context, tx *Tx, auth *ocs.Auth) error {
	auth.CreatedAt = tx.now
	auth.UpdatedAt = auth.CreatedAt

	if err := auth.Validate(); err != nil {
		return err
	}

	var expiry *string
	if auth.Expiry != nil {
		tmp := auth.Expiry.Format(time.RFC3339)
		expiry = &tmp
	}

	result, err := tx.ExecContext(ctx, `
    INSERT INTO auths (
      student_id,
      source,
      source_id,
      access_token,
      refresh_token,
      expiry,
      created_at,
      updated_at
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `,
		auth.StudentID,
		auth.Source,
		auth.SourceID,
		auth.AccessToken,
		auth.RefreshToken,
		expiry,
		(*NullTime)(&auth.CreatedAt),
		(*NullTime)(&auth.UpdatedAt),
	)
	if err != nil {
		return FormatError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	auth.ID = int(id)

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

func findAuthBySourceID(ctx context.Context, tx *Tx, source, sourceID string) (*ocs.Auth, error) {
	auths, _, err := findAuths(ctx, tx, ocs.AuthFilter{Source: &source, SourceID: &sourceID})
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

func updateAuth(ctx context.Context, tx *Tx, id int, accessToken, refreshToken string, expiry *time.Time) (*ocs.Auth, error) {
	auth, err := findAuthByID(ctx, tx, id)
	if err != nil {
		return auth, err
	}

	auth.AccessToken = accessToken
	auth.RefreshToken = refreshToken
	auth.Expiry = expiry
	auth.UpdatedAt = tx.now

	if err := auth.Validate(); err != nil {
		return auth, err
	}

	var expiryString *string
	if auth.Expiry != nil {
		v := auth.Expiry.Format(time.RFC3339)
		expiryString = &v
	}

	if _, err := tx.ExecContext(ctx, `
    UPDATE students
    SET access_token = ?
        refresh_token = ?
        expiry = ?
        updated_at = ?
    WHERE id = ?
  `,
		auth.AccessToken,
		auth.RefreshToken,
		expiryString,
		(*NullTime)(&auth.UpdatedAt),
		id,
	); err != nil {
		return auth, FormatError(err)
	}

	return auth, nil
}

func attachAuthAssociations(ctx context.Context, tx *Tx, auth *ocs.Auth) (err error) {
	if auth.Student, err = findStudentByID(ctx, tx, auth.StudentID); err != nil {
		return fmt.Errorf("attach auth user: %w", err)
	}
	return nil
}
