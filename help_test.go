package cli

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestShowHelpRootApp(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "my app"
	app.Version = "9.9.9"
	app.Description = "a test app"
	app.Subcommands = []*Command{
		{Name: "sub", Usage: "a subcommand", Action: func(context.Context, *Input, *Output) error { return nil }},
	}
	stdout, _, err := runApp(app, "--help")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, want := range []string{"NAME:", "USAGE:", "VERSION:", "9.9.9", "DESCRIPTION:", "COMMANDS:", "GLOBAL OPTIONS:"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help missing %q in:\n%s", want, stdout)
		}
	}
}

func TestShowHelpIncludesSubcommandsAndVersion(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Version = "1.0.0"
	app.Subcommands = []*Command{
		{Name: "alpha", Usage: "first"},
		{Name: "beta", Usage: "second"},
	}
	stdout, _, err := runApp(app, "--help")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout, "alpha") || !strings.Contains(stdout, "beta") {
		t.Errorf("subcommands missing in help:\n%s", stdout)
	}
	// The builtin version command should appear
	if !strings.Contains(stdout, "version") {
		t.Errorf("version subcommand missing in help:\n%s", stdout)
	}
}

func TestShowHelpSkipsHiddenCommand(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Subcommands = []*Command{
		{Name: "visible", Usage: "shown"},
		{Name: "secret", Usage: "hidden", Hidden: true},
	}
	stdout, _, err := runApp(app, "--help")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if strings.Contains(stdout, "secret") {
		t.Errorf("hidden command should not appear in help:\n%s", stdout)
	}
	if !strings.Contains(stdout, "visible") {
		t.Errorf("visible command missing in help:\n%s", stdout)
	}
}

func TestShowHelpSubcommand(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Subcommands = []*Command{
		{
			Name:        "sub",
			Usage:       "do sub",
			Description: "sub does things",
			Flags:       []Flag{&StringFlag{Name: "opt", Usage: "an option"}},
			Action:      func(context.Context, *Input, *Output) error { return nil },
		},
	}
	stdout, _, err := runApp(app, "sub", "--help")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, want := range []string{"NAME:", "USAGE:", "OPTIONS:", "--opt", "an option"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("sub help missing %q in:\n%s", want, stdout)
		}
	}
}

func TestShowHelpSubcommandUsageText(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Subcommands = []*Command{
		{
			Name:      "sub",
			Usage:     "do sub",
			UsageText: "prog sub [--flag] <name>",
			Action:    func(context.Context, *Input, *Output) error { return nil },
		},
	}
	stdout, _, err := runApp(app, "help", "sub")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout, "prog sub [--flag] <name>") {
		t.Errorf("UsageText not shown:\n%s", stdout)
	}
}

func TestShowHelpSubcommandNoFlagsSection(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Subcommands = []*Command{
		{Name: "sub", Usage: "no flags", Action: func(context.Context, *Input, *Output) error { return nil }},
	}
	stdout, _, err := runApp(app, "sub", "--help")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if strings.Contains(stdout, "OPTIONS:") {
		t.Errorf("should not show OPTIONS section without flags:\n%s", stdout)
	}
}

func TestShowHelpUnknownTopic(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Subcommands = []*Command{
		{Name: "sub", Usage: "sub", Action: func(context.Context, *Input, *Output) error { return nil }},
	}
	_, _, err := runApp(app, "help", "missing")
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestRenderFlagsHiddenAndAliases(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	o, out, _ := newBufOutput()
	app.renderFlags(o, []Flag{
		&StringFlag{Name: "visible", Aliases: []string{"v"}, Usage: "shown"},
		&BoolFlag{Name: "hidden", Hidden: true, Usage: "hidden"},
	}, true)
	s := out.String()
	if !strings.Contains(s, "--visible, -v") {
		t.Errorf("alias not shown:\n%s", s)
	}
	if strings.Contains(s, "hidden") {
		t.Errorf("hidden flag shown:\n%s", s)
	}
	if !strings.Contains(s, "--help, -h") {
		t.Errorf("builtin help not shown:\n%s", s)
	}
}

func TestShowHelpCommandHelpViaRunHelp(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Usage = "app"
	app.Subcommands = []*Command{
		{Name: "sub", Usage: "sub usage", Action: func(context.Context, *Input, *Output) error { return nil }},
	}
	stdout, _, err := runApp(app, "help")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout, "COMMANDS:") {
		t.Errorf("help output:\n%s", stdout)
	}
}
