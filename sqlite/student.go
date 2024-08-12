package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"github.com/maliByatzes/ocs"
)

var _ ocs.StudentService = (*StudentService)(nil)

type StudentService struct {
	db *DB
}

func NewStudentService(db *DB) *StudentService {
	return &StudentService{db: db}
}

func (s *StudentService) FindStudentByID(ctx context.Context, id int) (*ocs.Student, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	student, err := findStudentByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachStudentsAuths(ctx, tx, student); err != nil {
		return nil, err
	}

	return student, nil
}

func (s *StudentService) FindStudents(ctx context.Context, filter ocs.StudentFilter) ([]*ocs.Student, error) {
	return nil, nil
}

func (s *StudentService) CreateStudent(ctx context.Context, student *ocs.Student) error {
	return nil
}

func (s *StudentService) UpdateStudent(ctx context.Context, id int, upd ocs.StudentUpdate) (*ocs.Student, error) {
	return nil, nil
}

func (s *StudentService) DeleteStudent(ctx context.Context, id int) error {
	return nil
}

func findStudentByID(ctx context.Context, tx *Tx, id int) (*ocs.Student, error) {
	a, _, err := findStudents(ctx, tx, ocs.StudentFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, &ocs.Error{Code: ocs.ENOTFOUND, Message: "Student not found"}
	}
	return a[0], nil
}

func findStudents(ctx context.Context, tx *Tx, filter ocs.StudentFilter) (_ []*ocs.Student, n int, err error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Email; v != nil {
		where, args = append(where, "email = ?"), append(args, *v)
	}
	if v := filter.APIKey; v != nil {
		where, args = append(where, "api_key = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
    SELECT
      id,
      name,
      email,
      api_key,
      created_at,
      updated_at,
      COUNT(*) OVER()
    FROM students
    WHERE`+strings.Join(where, " AND ")+`
    ORDER BY id ASC
    `+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	students := make([]*ocs.Student, 0)
	for rows.Next() {
		var email sql.NullString
		var student ocs.Student
		if err := rows.Scan(
			&student.ID,
			&student.Name,
			&email,
			&student.APIKey,
			(*NullTime)(&student.CreatedAt),
			(*NullTime)(&student.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}

		if email.Valid {
			student.Email = email.String
		}

		students = append(students, &student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, n, nil
}

func attachStudentsAuths(ctx context.Context, tx *Tx, student *ocs.Student) (err error) {
	/*
		if student.Auths, _, err = findAuths(ctx, tx, ocs.AuthFilter{StudentID: &student.ID}); err != nil {
			return fmt.Errorf("attach student auths: %w", err)
		}*/
	return nil
}
