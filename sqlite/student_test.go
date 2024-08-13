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

    /*
		student2 := &ocs.Student{Name: "james"}
		if err := s.CreateStudent(context.Background(), student2); err != nil {
			t.Fatal(err)
		} else if got, want := student2.ID, 2; got != want {
			t.Fatalf("ID=%v, want=%v", got, want)
		}*/

		if other, err := s.FindStudentByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(student, other) {
			t.Fatalf("mismatch: %#v != %#v", student, other)
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
