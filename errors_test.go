package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestExitErrorMessage(t *testing.T) {
	e := &ExitError{Code: 5}
	if e.Error() != "exit status 5" {
		t.Errorf("Error = %q", e.Error())
	}
	if e.ExitCode() != 5 {
		t.Errorf("ExitCode = %d", e.ExitCode())
	}
}

func TestExitErrorWithErr(t *testing.T) {
	e := &ExitError{Code: 2, Err: errors.New("boom")}
	if e.Error() != "boom" {
		t.Errorf("Error = %q", e.Error())
	}
	if errors.Unwrap(e) == nil {
		t.Error("expected unwrap")
	}
}

func TestExitWithDetail(t *testing.T) {
	err := Exit(9, fmt.Sprintf("detail %d", 1))
	var ee *ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("Exit should return *ExitError, got %T", err)
	}
	if ee.Code != 9 || ee.Error() != "detail 1" {
		t.Errorf("Exit = %+v", ee)
	}
}

func TestExitWithoutDetail(t *testing.T) {
	err := Exit(3)
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 || ee.Err != nil {
		t.Fatalf("Exit = %+v", err)
	}
}

func TestUsageError(t *testing.T) {
	e := &UsageError{Msg: "bad usage"}
	if e.Error() != "bad usage" {
		t.Errorf("Error = %q", e.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	e := &NotFoundError{Command: "foo"}
	if !strings.Contains(e.Error(), "foo") {
		t.Errorf("Error = %q", e.Error())
	}
}
