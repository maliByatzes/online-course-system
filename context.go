package ocs

import "context"

type contextKey int

const (
	studentContextKey = contextKey(iota + 1)
	flashContextKey
)

func NewContextWithStudent(ctx context.Context, student *Student) context.Context {
	return context.WithValue(ctx, studentContextKey, student)
}

func StudentFromContext(ctx context.Context) *Student {
	student, _ := ctx.Value(studentContextKey).(*Student)
	return student
}

func StudentIDFromContext(ctx context.Context) int {
	if student := StudentFromContext(ctx); student != nil {
		return student.ID
	}
	return 0
}

func NewContextWithFlash(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, flashContextKey, v)
}

func FlashFromContext(ctx context.Context) string {
	v, _ := ctx.Value(flashContextKey).(string)
	return v
}
