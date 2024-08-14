package sqlite_test

import (
	"context"
	"testing"

	"github.com/maliByatzes/ocs"
	"github.com/maliByatzes/ocs/sqlite"
)

func TestAuthService_FindAuths(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)

		ctx := context.Background()

		MustCreateAuth(t, context.Background(), db, &ocs.Auth{
			Source:      "SRCA",
			SourceID:    "X1",
			AccessToken: "ACCESSX1",
			Student:     &ocs.Student{Name: "X", Email: "x@y.com"},
		})
		MustCreateAuth(t, context.Background(), db, &ocs.Auth{
			Source:      "SRCB",
			SourceID:    "X2",
			AccessToken: "ACCESSX2",
			Student:     &ocs.Student{Name: "X", Email: "x@y.com"},
		})
		MustCreateAuth(t, context.Background(), db, &ocs.Auth{
			Source:      ocs.AuthSourceGithub,
			SourceID:    "Y",
			AccessToken: "ACCESSY",
			Student:     &ocs.Student{Name: "Y", Email: "y@x.com"},
		})

		studentID := 1
		if a, n, err := s.FindAuths(ctx, ocs.AuthFilter{StudentID: &studentID}); err != nil {
			t.Fatal(err)
		} else if got, want := len(a), 2; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := a[0].SourceID, "X1"; got != want {
			t.Fatalf("[]=%v, want %v", got, want)
		} else if got, want := a[1].SourceID, "X2"; got != want {
			t.Fatalf("[]=%v, want %v", got, want)
		} else if got, want := n, 2; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

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

func MustCreateAuth(tb testing.TB, ctx context.Context, db *sqlite.DB, auth *ocs.Auth) (*ocs.Auth, context.Context) {
	tb.Helper()
	if err := sqlite.NewAuthService(db).CreateAuth(ctx, auth); err != nil {
		tb.Fatal(err)
	}
	return auth, ocs.NewContextWithStudent(ctx, auth.Student)
}
