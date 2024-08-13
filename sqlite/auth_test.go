package sqlite_test

import (
	"context"
	"testing"

	"github.com/maliByatzes/ocs"
	"github.com/maliByatzes/ocs/sqlite"
)

func TestAuthService_FindAuth(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)
		if _, err := s.FindAuthByID(context.Background(), 1); ocs.ErrorCode(err) != ocs.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
