package sqlite_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/maliByatzes/ocs"
	"github.com/maliByatzes/ocs/sqlite"
)

func TestStudentService_CreateStudent(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)

		student := &ocs.Student{
			Name:  "john",
			Email: "john@email.com",
		}

		if err := s.CreateStudent(context.Background(), student); err != nil {
			t.Fatal(err)
		} else if got, want := student.ID, 1; got != want {
			t.Fatalf("ID=%v, want=%v", got, want)
		} else if student.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		} else if student.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		student2 := &ocs.Student{Name: "james", Email: "james@email.com"}
		if err := s.CreateStudent(context.Background(), student2); err != nil {
			t.Fatal(err)
		} else if got, want := student2.ID, 2; got != want {
			t.Fatalf("ID=%v, want=%v", got, want)
		}

		if other, err := s.FindStudentByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(student, other) {
			t.Fatalf("mismatch: %#v != %#v", student, other)
		}
	})

	t.Run("ErrNameRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)
		if err := s.CreateStudent(context.Background(), &ocs.Student{}); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EINVALID || ocs.ErrorMessage(err) != `Provide required fields` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrEmailRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)
		if err := s.CreateStudent(context.Background(), &ocs.Student{Name: "kyle"}); err == nil {
			t.Fatalf("expected error")
		} else if ocs.ErrorCode(err) != ocs.EINVALID || ocs.ErrorMessage(err) != `Provide required fields` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestStudentService_UpdateStudent(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)

		student0, ctx0 := MustCreateStudent(t, context.Background(), db, &ocs.Student{
			Name:  "john",
			Email: "john@email.com",
		})

		newName, newEmail := "jeff", "jeff@email.com"
		us, err := s.UpdateStudent(ctx0, student0.ID, ocs.StudentUpdate{
			Name:  &newName,
			Email: &newEmail,
		})

		if err != nil {
			t.Fatal(err)
		} else if got, want := us.Name, newName; got != want {
			t.Fatalf("Name=%v, want %v", got, want)
		} else if got, want := us.Email, newEmail; got != want {
			t.Fatalf("Email=%v, want %v", got, want)
		}

		if other, err := s.FindStudentByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(us, other) {
			t.Fatalf("mismatch: %#v != %#v", us, other)
		}
	})

	t.Run("ErrUnauthorized", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)

		student0, _ := MustCreateStudent(t, context.Background(), db, &ocs.Student{Name: "joe", Email: "joe@email.com"})
		_, ctx1 := MustCreateStudent(t, context.Background(), db, &ocs.Student{Name: "kim", Email: "kim@email.com"})

		newName := "jeff"
		if _, err := s.UpdateStudent(ctx1, student0.ID, ocs.StudentUpdate{Name: &newName}); err == nil {
			t.Fatal("expected error")
		} else if ocs.ErrorCode(err) != ocs.EUNAUTHORIZED || ocs.ErrorMessage(err) != `You are not allowed to update this student.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestStudentService_FindStudent(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)
		if _, err := s.FindStudentByID(context.Background(), 1); ocs.ErrorCode(err) != ocs.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestStudentService_FindStudents(t *testing.T) {
	t.Run("Email", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewStudentService(db)

		ctx := context.Background()
		MustCreateStudent(t, ctx, db, &ocs.Student{Name: "john", Email: "john@email.com"})
		MustCreateStudent(t, ctx, db, &ocs.Student{Name: "jane", Email: "jane@email.com"})
		MustCreateStudent(t, ctx, db, &ocs.Student{Name: "frank", Email: "frank@email.com"})
		MustCreateStudent(t, ctx, db, &ocs.Student{Name: "joe", Email: "joe@email.com"})

		email := "joe@email.com"
		if a, n, err := s.FindStudents(ctx, ocs.StudentFilter{Email: &email}); err != nil {
			t.Fatal(err)
		} else if got, want := len(a), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := a[0].Name, "joe"; got != want {
			t.Fatalf("name=%v, want %v", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

func MustCreateStudent(tb testing.TB, ctx context.Context, db *sqlite.DB, student *ocs.Student) (*ocs.Student, context.Context) {
	tb.Helper()
	if err := sqlite.NewStudentService(db).CreateStudent(ctx, student); err != nil {
		tb.Fatal(err)
	}
	return student, ocs.NewContextWithStudent(ctx, student)
}
