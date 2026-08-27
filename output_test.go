package cli

import (
	"strings"
	"testing"
)

func TestOutputInfoAndVerbosity(t *testing.T) {
	o, out, _ := newBufOutput()
	o.Info("info message")
	o.Infof("formatted %s", "x")
	if !strings.Contains(out.String(), "info message") || !strings.Contains(out.String(), "formatted x") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestOutputQuietSuppressesInfo(t *testing.T) {
	o, out, _ := newBufOutput()
	o.SetVerbosity(VerbosityQuiet)
	o.Info("hidden")
	if out.String() != "" {
		t.Errorf("stdout should be empty, got %q", out.String())
	}
}

func TestOutputDebugRequiresVerbose(t *testing.T) {
	o, out, _ := newBufOutput()
	o.Debug("debug line")
	if out.String() != "" {
		t.Errorf("debug should not print at normal level: %q", out.String())
	}
	o.SetVerbosity(VerbosityDebug)
	o.Debug("now visible")
	o.Verbosef("verbose %d", 1)
	if !strings.Contains(out.String(), "now visible") || !strings.Contains(out.String(), "verbose 1") {
		t.Errorf("stdout = %q", out.String())
	}
}

func TestOutputWarnAndErrorGoToStderr(t *testing.T) {
	o, out, errb := newBufOutput()
	o.Warn("warn msg")
	o.Error("error msg")
	if out.String() != "" {
		t.Errorf("stdout should be empty, got %q", out.String())
	}
	if !strings.Contains(errb.String(), "warn msg") || !strings.Contains(errb.String(), "error msg") {
		t.Errorf("stderr = %q", errb.String())
	}
}

func TestOutputVerbosityAccessors(t *testing.T) {
	o := NewOutput()
	if o.GetVerbosity() != VerbosityNormal {
		t.Errorf("default verbosity = %d", o.GetVerbosity())
	}
	o.SetVerbosity(VerbosityQuiet)
	if o.GetVerbosity() != VerbosityQuiet {
		t.Errorf("verbosity = %d", o.GetVerbosity())
	}
}

func TestOutputTable(t *testing.T) {
	o, out, _ := newBufOutput()
	o.Table([]string{"A", "B"}, [][]string{{"1", "2"}, {"3", "4"}})
	s := out.String()
	if !strings.Contains(s, "A") || !strings.Contains(s, "1") || !strings.Contains(s, "2") {
		t.Errorf("table output = %q", s)
	}
}

func TestOutputRenderTableCustom(t *testing.T) {
	o, out, _ := newBufOutput()
	o.RenderTable(NewTable().SetStyle("box").SetHeader([]string{"X"}, AlignLeft).SetRows([][]string{{"1"}}, AlignLeft))
	if !strings.Contains(out.String(), "X") || !strings.Contains(out.String(), "1") {
		t.Errorf("custom table output = %q", out.String())
	}
}

func TestOutputTableAny(t *testing.T) {
	o, out, _ := newBufOutput()
	o.TableAny([]string{"N"}, [][]interface{}{{42}})
	if !strings.Contains(out.String(), "42") {
		t.Errorf("tableAny output = %q", out.String())
	}
}

func TestOutputAskNonInteractiveReturnsDefault(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = false
	o, _, _ := newBufOutput()
	o.root = app
	v, err := o.Ask("name", "default-value")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if v != "default-value" {
		t.Errorf("Ask = %v", v)
	}
}

func TestOutputAskInteractiveFromStdin(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("Zhang\n")
	o, _, _ := newBufOutput()
	o.root = app
	v, err := o.AskString("name", "world")
	if err != nil {
		t.Fatalf("AskString: %v", err)
	}
	if v != "Zhang" {
		t.Errorf("AskString = %q", v)
	}
}

func TestOutputConfirmInteractive(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("y\n")
	o, _, _ := newBufOutput()
	o.root = app
	got, err := o.Confirm("continue?", false)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !got {
		t.Error("Confirm should be true for y")
	}
}

func TestOutputChoiceInteractive(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("py\n")
	o, _, _ := newBufOutput()
	o.root = app
	got, err := o.Choice("lang", map[string]string{"go": "Go", "py": "Python"}, "go")
	if err != nil {
		t.Fatalf("Choice: %v", err)
	}
	if got != "py" {
		t.Errorf("Choice = %v", got)
	}
}

func TestOutputIsInteractiveRespectsNoInteraction(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.values = map[string]interface{}{"no-interaction": &boolValue{val: true}}
	o, _, _ := newBufOutput()
	o.root = app
	if o.isInteractive() {
		t.Error("no-interaction should disable")
	}
}

func TestOutputRemainingLogMethods(t *testing.T) {
	o, out, errb := newBufOutput()
	o.SetVerbosity(VerbosityDebug)
	o.Success("suc")
	o.Successf("suc %s", "fmt")
	o.Warnf("warn %s", "fmt")
	o.Errorf("err %s", "fmt")
	o.Debugf("dbg %s", "fmt")
	o.Verbose("verb")
	o.Verbosef("verb %s", "fmt")
	all := out.String() + "\n" + errb.String()
	for _, want := range []string{"suc", "suc fmt", "warn fmt", "err fmt", "dbg fmt", "verb", "verb fmt"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in output: %q", want, all)
		}
	}
}

func TestOutputAskHidden(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("p4ss\n")
	o, _, _ := newBufOutput()
	o.root = app
	got, err := o.AskHidden("password")
	if err != nil || got != "p4ss" {
		t.Fatalf("AskHidden = %q err=%v", got, err)
	}
}

func TestOutputAskHiddenNonInteractive(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.interactive = false
	o, _, _ := newBufOutput()
	o.root = app
	got, err := o.AskHidden("password")
	if err != nil || got != "" {
		t.Fatalf("AskHidden non-interactive = %q err=%v", got, err)
	}
}

func TestOutputAskHiddenWithoutStty(t *testing.T) {
	// Hidden-input branch without stty: read input directly.
	app := New()
	app.Name = "prog"
	app.interactive = true
	app.Stdin = strings.NewReader("plain\n")
	o, _, _ := newBufOutput()
	o.root = app
	q := NewQuestion("pwd", nil)
	q.SetHidden(true)
	v, err := NewAsk(o, q).Run()
	if err != nil || v != "plain" {
		t.Fatalf("hidden no-stty = %v err=%v", v, err)
	}
}
