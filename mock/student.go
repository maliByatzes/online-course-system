package mock

import (
	"context"

	"github.com/maliByatzes/ocs"
)

var _ ocs.StudentService = (*StudentService)(nil)

type StudentService struct {
	FindStudentByIDFn func(ctx context.Context, id int) (*ocs.Student, error)
	FindStudentsFn    func(ctx context.Context, filter ocs.StudentFilter) ([]*ocs.Student, int, error)
	CreateStudentFn   func(ctx context.Context, student *ocs.Student) error
	UpdateStudentFn   func(ctx context.Context, id int, upd *ocs.StudentUpdate) (*ocs.Student, error)
	DeleteStudentFn   func(ctx context.Context, id int) error
}

func (s *StudentService) FindStudentByID(ctx context.Context, id int) (*ocs.Student, error) {
	return s.FindStudentByID(ctx, id)
}

func (s *StudentService) FindStudents(ctx context.Context, filter ocs.StudentFilter) ([]*ocs.Student, int, error) {
	return s.FindStudents(ctx, filter)
}

func (s *StudentService) CreateStudent(ctx context.Context, student *ocs.Student) error {
	return s.CreateStudent(ctx, student)
}

func (s *StudentService) UpdateStudent(ctx context.Context, id int, upd ocs.StudentUpdate) (*ocs.Student, error) {
	return s.UpdateStudent(ctx, id, upd)
}

func (s *StudentService) DeleteStudent(ctx context.Context, id int) error {
	return s.DeleteStudent(ctx, id)
}
