package health

import "fmt"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.Resource == "" {
		return "resource not found"
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

type DatabaseError struct {
	Message string
	Cause   error
}

func (e *DatabaseError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "database error"
}

func (e *DatabaseError) Unwrap() error {
	return e.Cause
}
