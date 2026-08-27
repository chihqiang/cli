package cli

import (
	"errors"
	"fmt"
)

// ExitError carries an exit code and an error, for ending a command early with an exit code from an Action.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit status %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

// ExitCode returns the exit code.
func (e *ExitError) ExitCode() int { return e.Code }

// Exit builds an error carrying an exit code; usable in an Action as `return cli.Exit(1)`.
func Exit(code int, args ...interface{}) error {
	if len(args) > 0 {
		return &ExitError{Code: code, Err: errors.New(fmt.Sprint(args...))}
	}
	return &ExitError{Code: code}
}

// UsageError represents a command-line usage error (unknown option, missing required flag, etc.).
type UsageError struct {
	Msg string
}

func (e *UsageError) Error() string { return e.Msg }

// NotFoundError represents a command/help topic that does not exist.
type NotFoundError struct {
	Command string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("No help topic for %q", e.Command)
}
