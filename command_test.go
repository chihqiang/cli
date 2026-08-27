package cli

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// newCommandApp builds a test App with a subcommand.
func newCommandApp() *Command {
	app := New()
	app.Name = "prog"
	app.Usage = "test app"
	app.Version = "1.2.3"
	app.Subcommands = []*Command{
		{
			Name:    "greet",
			Aliases: []string{"g"},
			Usage:   "greet someone",
			Flags: []Flag{
				&StringFlag{Name: "name", Value: "World", Usage: "name to greet"},
				&BoolFlag{Name: "shout", Aliases: []string{"s"}, Usage: "shout"},
			},
			Action: func(ctx context.Context, in *Input, out *Output) error {
				out.Infof("hello %s", in.String("name"))
				if in.Bool("shout") {
					out.Infof("shouted")
				}
				return nil
			},
		},
	}
	return app
}

func TestRunSubcommandDispatch(t *testing.T) {
	app := newCommandApp()
	stdout, _, err := runApp(app, "greet", "--name", "Alice")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout, "hello Alice") {
		t.Errorf("expected greeting output, got %q", stdout)
	}
	if strings.Contains(stdout, "shouted") {
		t.Errorf("did not expect shout output, got %q", stdout)
	}
}

func TestRunSubcommandAlias(t *testing.T) {
	app := newCommandApp()
	stdout, _, err := runApp(app, "g", "--shout")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout, "hello World") || !strings.Contains(stdout, "shouted") {
		t.Errorf("alias run output mismatch: %q", stdout)
	}
}

func TestRunPositionalArgs(t *testing.T) {
	app := New()
	app.Name = "prog"
	var got Args
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		got = in.Args()
		return nil
	}
	_, _, err := runApp(app, "a", "b", "c")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got.Len() != 3 || got.First() != "a" || got.Get(2) != "c" {
		t.Errorf("unexpected args: %v", got.Slice())
	}
}

func TestRunHelpFlagBypassesRequired(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{&StringFlag{Name: "token", Required: true}}
	app.Action = func(ctx context.Context, in *Input, out *Output) error { return nil }
	stdout, _, err := runApp(app, "--help")
	if err != nil {
		t.Fatalf("--help should not error: %v", err)
	}
	if !strings.Contains(stdout, "USAGE:") {
		t.Errorf("expected help output, got %q", stdout)
	}
}

func TestRunMissingRequiredFlag(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{&StringFlag{Name: "token", Required: true}}
	app.Action = func(ctx context.Context, in *Input, out *Output) error { return nil }
	_, _, err := runApp(app)
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UsageError, got %v", err)
	}
}

func TestRunVersionFlagDefault(t *testing.T) {
	app := newCommandApp()
	stdout, _, err := runApp(app, "--version")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout, "1.2.3") {
		t.Errorf("expected version output, got %q", stdout)
	}
}

func TestRunVersionPrinterCustom(t *testing.T) {
	app := newCommandApp()
	app.VersionPrinter = func(ctx context.Context, in *Input, out *Output) {
		out.Infof("custom version %s", in.Command().Version)
	}
	stdout, _, err := runApp(app, "version")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout, "custom version 1.2.3") {
		t.Errorf("expected custom version output, got %q", stdout)
	}
}

func TestRunUnknownCommandWithHandler(t *testing.T) {
	app := newCommandApp()
	var gotCmd string
	app.CommandNotFound = func(ctx context.Context, in *Input, out *Output, cmd string) {
		gotCmd = cmd
		out.Errorf("not found: %s", cmd)
	}
	_, stderr, err := runApp(app, "nope")
	if err != nil {
		t.Fatalf("handler should swallow error: %v", err)
	}
	if gotCmd != "nope" {
		t.Errorf("expected nope, got %s", gotCmd)
	}
	if !strings.Contains(stderr, "not found: nope") {
		t.Errorf("expected stderr message, got %q", stderr)
	}
}

func TestRunUnknownCommandReturnsNotFound(t *testing.T) {
	app := newCommandApp()
	_, _, err := runApp(app, "nope")
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestRunNoActionShowsHelp(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "group app"
	app.Subcommands = []*Command{{Name: "sub", Usage: "a subcommand", Action: func(context.Context, *Input, *Output) error { return nil }}}
	stdout, _, err := runApp(app)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout, "COMMANDS:") {
		t.Errorf("expected app help, got %q", stdout)
	}
}

func TestGlobalBeforeAfterHooks(t *testing.T) {
	app := newCommandApp()
	var order []string
	app.Before = func(ctx context.Context, in *Input, out *Output) error {
		order = append(order, "app-before")
		return nil
	}
	app.After = func(ctx context.Context, in *Input, out *Output) error {
		order = append(order, "app-after")
		return nil
	}
	_, _, err := runApp(app, "greet")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	want := []string{"app-before", "app-after"}
	if len(order) != 2 || order[0] != want[0] || order[1] != want[1] {
		t.Errorf("hook order = %v, want %v", order, want)
	}
}

func TestCommandBeforeAfterOrder(t *testing.T) {
	app := New()
	app.Name = "prog"
	var order []string
	app.Subcommands = []*Command{
		{
			Name: "sub",
			Before: func(ctx context.Context, in *Input, out *Output) error {
				order = append(order, "before")
				return nil
			},
			Action: func(ctx context.Context, in *Input, out *Output) error {
				order = append(order, "action")
				return nil
			},
			After: func(ctx context.Context, in *Input, out *Output) error {
				order = append(order, "after")
				return nil
			},
		},
	}
	_, _, err := runApp(app, "sub")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if strings.Join(order, ",") != "before,action,after" {
		t.Errorf("order = %v", order)
	}
}

func TestBeforeErrorAborts(t *testing.T) {
	app := New()
	app.Name = "prog"
	called := false
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		called = true
		return nil
	}
	app.Before = func(ctx context.Context, in *Input, out *Output) error {
		return &ExitError{Code: 3, Err: errors.New("boom")}
	}
	_, _, err := runApp(app)
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("expected ExitError code 3, got %v", err)
	}
	if called {
		t.Error("action should not run after before error")
	}
}

func TestActionErrorPropagates(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		return errors.New("action boom")
	}
	_, _, err := runApp(app)
	if err == nil || err.Error() != "action boom" {
		t.Fatalf("expected action error, got %v", err)
	}
}

func TestCommandReusable(t *testing.T) {
	app := newCommandApp()
	for i := 0; i < 2; i++ {
		stdout, _, err := runApp(app, "greet", "--name", "Bob")
		if err != nil {
			t.Fatalf("run %d failed: %v", i, err)
		}
		if !strings.Contains(stdout, "hello Bob") {
			t.Errorf("run %d output mismatch: %q", i, stdout)
		}
	}
}

func TestFindSubcommandByNameAndAlias(t *testing.T) {
	app := newCommandApp()
	if got := app.findSubcommand("greet"); got == nil || got.Name != "greet" {
		t.Errorf("findSubcommand(greet) = %v", got)
	}
	if got := app.findSubcommand("g"); got == nil {
		t.Errorf("findSubcommand(g) should match alias")
	}
	if got := app.findSubcommand("missing"); got != nil {
		t.Errorf("findSubcommand(missing) = %v, want nil", got)
	}
}

func TestCommandNamesIncludesAliases(t *testing.T) {
	c := &Command{Name: "main", Aliases: []string{"a", "b"}}
	if c.Name != "main" || len(c.Aliases) != 2 || c.Aliases[0] != "a" {
		t.Errorf("command metadata mismatch: %+v", c)
	}
}

func TestCommandHasFlag(t *testing.T) {
	app := New()
	app.Flags = []Flag{&BoolFlag{Name: "verbose"}}
	if !app.hasFlag("verbose") {
		t.Error("hasFlag should detect user-defined verbose flag")
	}
	app2 := New()
	app2.Flags = []Flag{&StringFlag{Name: "name"}}
	if app2.hasFlag("verbose") {
		t.Error("hasFlag should be false for undeclared flag")
	}
}

func TestAppTypeAlias(t *testing.T) {
	var a *App = New()
	if a == nil {
		t.Fatal("App alias should construct")
	}
}

func TestExitAndExitError(t *testing.T) {
	err := Exit(7, "detail")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 7 || ee.ExitCode() != 7 {
		t.Fatalf("Exit(7) = %v", err)
	}
	if ee.Error() != "detail" {
		t.Errorf("Error = %q", ee.Error())
	}
	if errors.Unwrap(ee) == nil {
		t.Error("expected unwrap chain")
	}
}

func TestOutputPropagatesFromRoot(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Subcommands = []*Command{
		{
			Name: "sub",
			Action: func(ctx context.Context, in *Input, out *Output) error {
				out.Info("shared output")
				return nil
			},
		},
	}
	stdout, _, err := runApp(app, "sub")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout, "shared output") {
		t.Errorf("expected output, got %q", stdout)
	}
}

func TestOutputExistingInstanceReused(t *testing.T) {
	// output() creates and attaches root when uninitialized; returns the same instance when it already exists.
	c := &Command{}
	o1 := c.output()
	o2 := c.output()
	if o1 != o2 {
		t.Error("output() should return the same instance")
	}
	if c.out == nil || c.out.root != c {
		t.Errorf("output() should set root: out=%v root=%v", c.out, c.out.root)
	}
	// When app.out is already set, subcommands reuse app.out.
	app := New()
	app.Name = "prog"
	o, _, _ := newBufOutput()
	app.out = o
	sub := &Command{Name: "sub"}
	sub.app = app
	if got := sub.output(); got != o {
		t.Error("subcommand should reuse app output")
	}
}
