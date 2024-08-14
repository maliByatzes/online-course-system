package sqlite_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/maliByatzes/ocs"
	"github.com/maliByatzes/ocs/sqlite"
)

func TestAuthService_CreateAuth(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)

		expiry := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
		auth := &ocs.Auth{
			Source:       ocs.AuthSourceGithub,
			SourceID:     "SOURCEID",
			AccessToken:  "ACCESS",
			RefreshToken: "REFRESH",
			Expiry:       &expiry,
			Student: &ocs.Student{
				Name:  "jeff",
				Email: "jeff@email.com",
			},
		}

		if err := s.CreateAuth(context.Background(), auth); err != nil {
			t.Fatal(err)
		} else if got, want := auth.ID, 1; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		} else if auth.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		} else if auth.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		if other, err := s.FindAuthByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(other, auth) {
			t.Fatalf("mismatch: %#v != %#v", auth, other)
		}

		if student, err := sqlite.NewStudentService(db).FindStudentByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if len(student.Auths) != 1 {
			t.Fatal("expected auths")
		} else if auth := student.Auths[0]; auth.ID != 1 {
			t.Fatalf("unexpected auth: %#v", auth)
		}
	})

	t.Run("ErrSourceRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		if err := sqlite.NewAuthService(db).CreateAuth(context.Background(), &ocs.Auth{
			Student: &ocs.Student{Name: "NAME", Email: "NAME@EMAIL.COM"},
		}); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EINVALID || ocs.ErrorMessage(err) != `Source required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrSourceIDRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		if err := sqlite.NewAuthService(db).CreateAuth(context.Background(), &ocs.Auth{
			Source:  ocs.AuthSourceGithub,
			Student: &ocs.Student{Name: "NAME", Email: "NAME@EMAIL.COM"},
		}); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EINVALID || ocs.ErrorMessage(err) != `Source ID required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrAccessTokenRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)
		if err := s.CreateAuth(context.Background(), &ocs.Auth{
			Source:   ocs.AuthSourceGithub,
			SourceID: "X",
			Student:  &ocs.Student{Name: "NAME", Email: "NAME@EMAIL.COM"},
		}); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EINVALID || ocs.ErrorMessage(err) != `Access Token required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrStudentRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)
		if err := s.CreateAuth(context.Background(), &ocs.Auth{}); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EINVALID || ocs.ErrorMessage(err) != `Student required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestAuthService_DeleteAuth(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)
		auth0, ctx0 := MustCreateAuth(t, context.Background(), db, &ocs.Auth{
			Source:      ocs.AuthSourceGithub,
			SourceID:    "X",
			AccessToken: "X",
			Student: &ocs.Student{
				Name:  "X",
				Email: "X@Y.COM",
			},
		})

		if err := s.DeleteAuth(ctx0, auth0.ID); err != nil {
			t.Fatal(err)
		} else if _, err := s.FindAuthByID(ctx0, auth0.ID); ocs.ErrorCode(err) != ocs.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)
		if err := s.DeleteAuth(context.Background(), 1); ocs.ErrorCode(err) != ocs.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAuthService(db)

		auth0, _ := MustCreateAuth(t, context.Background(), db, &ocs.Auth{
			Source:      ocs.AuthSourceGithub,
			SourceID:    "X",
			AccessToken: "X", Student: &ocs.Student{Name: "X", Email: "X@EMAIL.COM"},
		})
		_, ctx1 := MustCreateAuth(t, context.Background(), db, &ocs.Auth{
			Source:      ocs.AuthSourceGithub,
			SourceID:    "Y",
			AccessToken: "Y", Student: &ocs.Student{Name: "Y", Email: "Y@EMAIL.COM"},
		})

		if err := s.DeleteAuth(ctx1, auth0.ID); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EUNAUTHORIZED || ocs.ErrorMessage(err) != `You are not allowed to delete this auth.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestAuthService_FindAuths(t *testing.T) {
	t.Run("Student", func(t *testing.T) {
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
