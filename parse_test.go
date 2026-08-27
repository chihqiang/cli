package cli

import (
	"errors"
	"strings"
	"testing"
)

// addBool registers a bool flag in defs.
func addBool(defs map[string]*flagDef, name string) {
	b := &boolValue{}
	defs[name] = &flagDef{value: b, isBool: true}
}

func TestParseTokensLongForm(t *testing.T) {
	defs := map[string]*flagDef{
		"name": {value: &singleValue{}, isBool: false},
	}
	set := map[string]bool{}
	rest, err := parseTokens([]string{"--name", "Alice", "--name=Bob"}, defs, set, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rest) != 0 {
		t.Errorf("rest = %v", rest)
	}
	if !set["name"] {
		t.Error("name should be set")
	}
}

func TestParseTokensShortForm(t *testing.T) {
	sv := &singleValue{}
	defs := map[string]*flagDef{
		"name": {value: sv, isBool: false},
		"n":    {value: sv, isBool: false},
	}
	set := map[string]bool{}
	rest, err := parseTokens([]string{"-n", "Alice"}, defs, set, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rest) != 0 || !set["n"] || sv.String() != "Alice" {
		t.Errorf("rest=%v set=%v value=%q", rest, set, sv.String())
	}
}

func TestParseTokensBoolFlag(t *testing.T) {
	defs := map[string]*flagDef{}
	addBool(defs, "force")
	set := map[string]bool{}
	// A bool flag does not consume the next token
	rest, err := parseTokens([]string{"--force", "arg1"}, defs, set, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rest) != 1 || rest[0] != "arg1" {
		t.Errorf("rest = %v", rest)
	}
	if !set["force"] {
		t.Error("force should be set")
	}
}

func TestParseTokensInterspersedFalse(t *testing.T) {
	defs := map[string]*flagDef{
		"name": {value: &singleValue{}, isBool: false},
	}
	set := map[string]bool{}
	// interspersed=false: stops parsing flags after the first positional argument
	rest, err := parseTokens([]string{"cmd", "--name", "x"}, defs, set, false)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rest) != 3 {
		t.Errorf("rest = %v", rest)
	}
	if set["name"] {
		t.Error("name should not be parsed after positional")
	}
}

func TestParseTokensDoubleDash(t *testing.T) {
	defs := map[string]*flagDef{
		"name": {value: &singleValue{}, isBool: false},
	}
	set := map[string]bool{}
	rest, err := parseTokens([]string{"--", "--name", "x"}, defs, set, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rest) != 2 || rest[0] != "--name" {
		t.Errorf("rest = %v", rest)
	}
	if set["name"] {
		t.Error("name should not be parsed after --")
	}
}

func TestParseTokensUnknownFlag(t *testing.T) {
	defs := map[string]*flagDef{}
	_, err := parseTokens([]string{"--nope"}, defs, map[string]bool{}, true)
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UsageError, got %v", err)
	}
	if !strings.Contains(ue.Msg, "nope") {
		t.Errorf("msg = %q", ue.Msg)
	}
}

func TestParseTokensMissingValue(t *testing.T) {
	defs := map[string]*flagDef{
		"name": {value: &singleValue{}, isBool: false},
	}
	_, err := parseTokens([]string{"--name"}, defs, map[string]bool{}, true)
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UsageError, got %v", err)
	}
}

func TestParseTokensInvalidValue(t *testing.T) {
	b := &boolValue{}
	defs := map[string]*flagDef{"ok": {value: b, isBool: false}}
	_, err := parseTokens([]string{"--ok", "notabool"}, defs, map[string]bool{}, true)
	if err == nil {
		t.Fatal("expected error for invalid bool value")
	}
}

func TestParseTokensDanglingDash(t *testing.T) {
	defs := map[string]*flagDef{}
	rest, err := parseTokens([]string{"-"}, defs, map[string]bool{}, true)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rest) != 1 || rest[0] != "-" {
		t.Errorf("rest = %v", rest)
	}
}

func TestSplitFlagToken(t *testing.T) {
	name, val, hasEq := splitFlagToken("--name=Alice")
	if name != "name" || val != "Alice" || !hasEq {
		t.Errorf("split long with eq: %q %q %v", name, val, hasEq)
	}
	name, val, hasEq = splitFlagToken("--force")
	if name != "force" || val != "true" || hasEq {
		t.Errorf("split long no eq: %q %q %v", name, val, hasEq)
	}
	name, val, hasEq = splitFlagToken("-n=x")
	if name != "n" || val != "x" || !hasEq {
		t.Errorf("split short with eq: %q %q %v", name, val, hasEq)
	}
	name, val, hasEq = splitFlagToken("-s")
	if name != "s" || val != "true" || hasEq {
		t.Errorf("split short no eq: %q %q %v", name, val, hasEq)
	}
}

func TestIsBoolFlag(t *testing.T) {
	if !isBoolFlag(&boolValue{}) {
		t.Error("boolValue should report IsBoolFlag")
	}
	if isBoolFlag(&singleValue{}) {
		t.Error("singleValue should not report IsBoolFlag")
	}
}
