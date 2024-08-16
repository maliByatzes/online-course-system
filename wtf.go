package ocs

import "context"

var (
	Version string
	Commit  string
)

var ReportError = func(ctx context.Context, err error, args ...interface{}) {}

var ReportPanic = func(err interface{}) {}
