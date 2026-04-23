package apperrors

import (
	stderrors "errors"
	"fmt"
	"strings"
)

const (
	ExitSuccess  = 0
	ExitGeneric  = 1
	ExitUsage    = 2
	ExitAuth     = 3
	ExitScope    = 4
	ExitNotFound = 5
	ExitNetwork  = 6
	ExitServer   = 7
)

type Error struct {
	Code    int
	Message string
	Err     error
	Details []string
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) ExitCode() int {
	if e.Code == 0 {
		return ExitGeneric
	}
	return e.Code
}

type exitCoder interface {
	ExitCode() int
}

type statusCoder interface {
	HTTPStatusCode() int
	ErrorDetail() string
	ErrorDetails() []string
}

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func WithDetails(err error, details ...string) error {
	var target *Error
	if stderrors.As(err, &target) {
		target.Details = append(target.Details, details...)
		return target
	}
	return &Error{Code: ExitGeneric, Message: err.Error(), Err: err, Details: details}
}

func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var codeErr exitCoder
	if stderrors.As(err, &codeErr) {
		return codeErr.ExitCode()
	}

	var statusErr statusCoder
	if stderrors.As(err, &statusErr) {
		switch status := statusErr.HTTPStatusCode(); {
		case status == 401:
			return ExitAuth
		case status == 403:
			return ExitScope
		case status == 404:
			return ExitNotFound
		case status >= 500:
			return ExitServer
		case status >= 400:
			return ExitUsage
		default:
			return ExitGeneric
		}
	}

	return ExitGeneric
}

func Format(err error) string {
	if err == nil {
		return ""
	}

	var statusErr statusCoder
	if stderrors.As(err, &statusErr) {
		parts := []string{statusErr.ErrorDetail()}
		parts = append(parts, statusErr.ErrorDetails()...)
		return strings.Join(filterEmpty(parts), "\n")
	}

	var appErr *Error
	if stderrors.As(err, &appErr) {
		parts := []string{appErr.Message}
		parts = append(parts, appErr.Details...)
		return strings.Join(filterEmpty(parts), "\n")
	}

	return err.Error()
}

func filterEmpty(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			out = append(out, item)
		}
	}
	return out
}

func MissingFlag(name string) error {
	return New(ExitUsage, fmt.Sprintf("missing required flag: --%s", name))
}
