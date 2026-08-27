package cli

import (
	"testing"
	"time"
)

// newTestInput builds an Input with several flag underlying values (injecting private fields directly in the same package).
func newTestInput() *Input {
	cmd := New()
	cmd.Name = "cmd"
	cmd.values = map[string]interface{}{
		"name":   &singleValue{val: "Go"},
		"count":  &singleValue{val: "3"},
		"big":    &singleValue{val: "9000000000"},
		"small":  &singleValue{val: "7"},
		"ratio":  &singleValue{val: "0.25"},
		"dur":    &singleValue{val: "2s"},
		"ok":     &boolValue{val: true},
		"off":    &boolValue{val: false},
		"tags":   &sliceValue{vals: []string{"a", "b"}},
		"nums":   &sliceValue{vals: []string{"1", "2"}},
		"floats": &sliceValue{vals: []string{"0.1", "0.2"}},
	}
	return &Input{cmd: cmd}
}

func TestArgsMethods(t *testing.T) {
	a := Args{"x", "y", "z"}
	if a.First() != "x" || a.Get(1) != "y" || a.Len() != 3 || !a.Present() {
		t.Errorf("Args methods mismatch: %v", a.Slice())
	}
	if len(a.Slice()) != 3 || len(a.Tail()) != 2 || a.Tail()[0] != "y" {
		t.Errorf("Tail = %v", a.Tail())
	}
	empty := Args{}
	if empty.First() != "" || empty.Present() || len(empty.Tail()) != 0 {
		t.Errorf("empty Args should be empty")
	}
	single := Args{"only"}
	if len(single.Tail()) != 0 {
		t.Errorf("single Tail should be empty")
	}
}

func TestInputStringAndBool(t *testing.T) {
	in := newTestInput()
	if in.String("name") != "Go" {
		t.Errorf("String = %q", in.String("name"))
	}
	if !in.Bool("ok") || in.Bool("off") {
		t.Error("Bool mismatch")
	}
	if in.String("missing") != "" {
		t.Error("missing flag should return empty")
	}
}

func TestFindCommand(t *testing.T) {
	root := New()
	root.Name = "app"
	build := &Command{Name: "build", Aliases: []string{"b"}}
	inner := &Command{Name: "inner"}
	build.Subcommands = append(build.Subcommands, inner)
	root.Subcommands = append(root.Subcommands, build)
	root.Subcommands = append(root.Subcommands, &Command{Name: "test"})
	root.app = root

	in := &Input{cmd: root}
	if in.FindCommand("build") != build {
		t.Error("FindCommand(build) should find build")
	}
	if in.FindCommand("b") != build {
		t.Error("FindCommand(b) should find build via alias")
	}
	if in.FindCommand("inner") != inner {
		t.Error("FindCommand(inner) should find nested subcommand")
	}
	if in.FindCommand("missing") != nil {
		t.Error("FindCommand(missing) should be nil")
	}

	// With no bound root command (app is nil), returns nil without panicking.
	nilIn := &Input{cmd: New()}
	if nilIn.FindCommand("build") != nil {
		t.Error("FindCommand without app should be nil")
	}
}

func TestInputTypedGetters(t *testing.T) {
	in := newTestInput()
	if in.Int("count") != 3 {
		t.Errorf("Int = %d", in.Int("count"))
	}
	if in.Int64("big") != 9000000000 {
		t.Errorf("Int64 = %d", in.Int64("big"))
	}
	if in.Uint("small") != 7 {
		t.Errorf("Uint = %d", in.Uint("small"))
	}
	if in.Float64("ratio") != 0.25 {
		t.Errorf("Float64 = %v", in.Float64("ratio"))
	}
	if in.Duration("dur") != 2*time.Second {
		t.Errorf("Duration = %v", in.Duration("dur"))
	}
	if in.Int("missing") != 0 {
		t.Errorf("missing Int should be 0")
	}
}

func TestInputSliceGetters(t *testing.T) {
	in := newTestInput()
	if got := in.StringSlice("tags"); len(got) != 2 || got[0] != "a" {
		t.Errorf("StringSlice = %v", got)
	}
	if got := in.IntSlice("nums"); len(got) != 2 || got[1] != 2 {
		t.Errorf("IntSlice = %v", got)
	}
	if got := in.Float64Slice("floats"); len(got) != 2 || got[0] != 0.1 {
		t.Errorf("Float64Slice = %v", got)
	}
	if got := in.StringSlice("name"); len(got) != 1 || got[0] != "Go" {
		t.Errorf("single flag via StringSlice = %v", got)
	}
	if in.StringSlice("missing") != nil {
		t.Error("missing slice should be nil")
	}
}

func TestInputGenericAndIsSet(t *testing.T) {
	in := newTestInput()
	if in.Generic("name") == nil {
		t.Error("Generic should return value")
	}
	if in.IsSet("name") {
		t.Error("IsSet should be false when not explicitly set")
	}
	in.cmd.setFlags = map[string]bool{"name": true}
	if !in.IsSet("name") {
		t.Error("IsSet should be true after explicit set")
	}
}

func TestInputCommandAndRoot(t *testing.T) {
	app := New()
	app.Name = "app"
	sub := &Command{Name: "sub"}
	sub.app = app
	in := &Input{cmd: sub}
	if in.Command().Name != "sub" {
		t.Errorf("Command = %v", in.Command().Name)
	}
	if in.Root() != app {
		t.Error("Root should be app")
	}
}

func TestInputIsInteractive(t *testing.T) {
	app := New()
	app.Name = "app"
	app.interactive = true
	app.values = map[string]interface{}{}
	app.setFlags = map[string]bool{}
	sub := &Command{Name: "sub"}
	sub.parent = app
	in := &Input{cmd: sub}
	if !in.IsInteractive() {
		t.Error("should be interactive")
	}
	// --no-interaction disables interaction
	app.values["no-interaction"] = &boolValue{val: true}
	if in.IsInteractive() {
		t.Error("no-interaction should disable interactivity")
	}
}

func TestInputLookupParentChain(t *testing.T) {
	app := New()
	app.Name = "app"
	app.values = map[string]interface{}{"global": &singleValue{val: "g"}}
	sub := &Command{Name: "sub"}
	sub.parent = app
	sub.values = map[string]interface{}{"local": &singleValue{val: "l"}}
	in := &Input{cmd: sub}
	if in.String("global") != "g" {
		t.Errorf("parent flag = %q", in.String("global"))
	}
	if in.String("local") != "l" {
		t.Errorf("local flag = %q", in.String("local"))
	}
}

func TestStringViaBoolValue(t *testing.T) {
	in := newTestInput()
	if in.String("ok") != "true" || in.String("off") != "false" {
		t.Errorf("bool via String: %q %q", in.String("ok"), in.String("off"))
	}
}

func TestInputNoArgs(t *testing.T) {
	in := &Input{cmd: New()}
	if in.NArg() != 0 || in.Args().Present() {
		t.Error("empty input should have no args")
	}
}
