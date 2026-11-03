package cli

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"
	"time"
)

func TestEnvVarsLookup(t *testing.T) {
	t.Setenv("CLI_TEST_A", "a")
	src := EnvVars("CLI_TEST_MISSING", "CLI_TEST_A")
	v, ok := src.Lookup()
	if !ok || v != "a" {
		t.Fatalf("Lookup = %q, %v; want a, true", v, ok)
	}
}

func TestEnvVarsLookupNone(t *testing.T) {
	src := EnvVars("CLI_TEST_NOPE_1", "CLI_TEST_NOPE_2")
	if _, ok := src.Lookup(); ok {
		t.Fatal("expected no match")
	}
}

func TestEnvVarsString(t *testing.T) {
	src := EnvVars("A", "B")
	if got := src.String(); got != "$A, $B" {
		t.Errorf("String = %q", got)
	}
}

func TestEnvVarsEmpty(t *testing.T) {
	src := EnvVars()
	if got := src.String(); got != "" {
		t.Errorf("String = %q, want empty", got)
	}
}

type fakeSource struct{ v string }

func (f fakeSource) Lookup() (string, bool) { return f.v, true }
func (f fakeSource) String() string         { return "fake" }

func TestCustomValueSource(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{&StringFlag{Name: "s", Sources: fakeSource{v: "from-fake"}}}
	var got string
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		got = in.String("s")
		return nil
	}
	_, _, err := runApp(app)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got != "from-fake" {
		t.Errorf("source value = %q", got)
	}
}

func TestStringFlagBasics(t *testing.T) {
	f := StringFlag{Name: "lang", Aliases: []string{"l"}, Value: "en", Required: true}
	if len(f.Names()) != 2 || f.Names()[0] != "lang" {
		t.Errorf("Names = %v", f.Names())
	}
	if !f.TakesValue() {
		t.Error("StringFlag should take value")
	}
	if !f.IsRequired() {
		t.Error("should be required")
	}
}

func TestBoolFlagDoesNotTakeValue(t *testing.T) {
	f := BoolFlag{Name: "force"}
	if f.TakesValue() {
		t.Error("BoolFlag should not take value")
	}
}

func TestFlagUsageTextDefaultAndSource(t *testing.T) {
	f := StringFlag{Name: "lang", Usage: "language", Value: "en", Sources: EnvVars("LANG")}
	u := f.common().usageText(f.Value)
	if !strings.Contains(u, "language") || !strings.Contains(u, "default: en") || !strings.Contains(u, "$LANG") {
		t.Errorf("usage = %q", u)
	}
}

func TestFlagUsageTextDefaultTextOverride(t *testing.T) {
	f := StringFlag{Name: "x", Usage: "opt", DefaultText: "42"}
	u := f.common().usageText("")
	if !strings.Contains(u, "default: 42") {
		t.Errorf("usage = %q", u)
	}
}

func TestFlagHidden(t *testing.T) {
	f := StringFlag{Name: "secret", Hidden: true}
	if !f.isHidden() {
		t.Error("should be hidden")
	}
	f2 := StringFlag{Name: "plain"}
	if f2.isHidden() {
		t.Error("should not be hidden")
	}
}

func TestSingleValueSetOverwrites(t *testing.T) {
	v := &singleValue{val: "default"}
	if err := v.Set("new"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if v.val != "new" || v.String() != "new" {
		t.Errorf("singleValue = %q", v.String())
	}
}

func TestSliceValueAppendsAndSplits(t *testing.T) {
	v := &sliceValue{vals: []string{"d"}}
	if err := v.Set("a,b"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("c"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := v.String(); got != "a,b,c" {
		t.Errorf("sliceValue = %q", got)
	}
}

func TestSliceValueFirstSetClearsDefault(t *testing.T) {
	v := &sliceValue{vals: []string{"default"}}
	_ = v.Set("x")
	if len(v.vals) != 1 || v.vals[0] != "x" {
		t.Errorf("vals = %v, want [x]", v.vals)
	}
}

func TestBoolValueSet(t *testing.T) {
	v := &boolValue{}
	if err := v.Set("true"); err != nil || !v.val || !v.isSet {
		t.Fatalf("boolValue set: %+v err=%v", v, err)
	}
	if v.String() != "true" {
		t.Errorf("boolValue.String = %q", v.String())
	}
	if !v.IsBoolFlag() {
		t.Error("boolValue should implement IsBoolFlag")
	}
	if err := v.Set("nonsense"); err == nil {
		t.Error("expected error for invalid bool")
	}
}

func TestRegisterTypedFlags(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{
		&StringFlag{Name: "s", Value: "str"},
		&IntFlag{Name: "i", Value: 3},
		&Int64Flag{Name: "i64", Value: 9},
		&UintFlag{Name: "u", Value: 2},
		&Float64Flag{Name: "f", Value: 1.5},
		&BoolFlag{Name: "b", Value: true},
		&DurationFlag{Name: "d", Value: 2 * time.Second},
		&StringSliceFlag{Name: "ss", Value: []string{"a"}},
		&IntSliceFlag{Name: "is", Value: []int{1}},
		&Float64SliceFlag{Name: "fs", Value: []float64{0.5}},
	}
	var captured string
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		captured = in.String("s") + in.String("i") + in.String("i64") + in.String("u") +
			in.String("f") + in.String("b") + in.String("d") + in.String("ss") + in.String("is") + in.String("fs")
		return nil
	}
	_, _, err := runApp(app)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(captured, "str") || !strings.Contains(captured, "3") {
		t.Errorf("captured = %q", captured)
	}
}

func TestGenericFlagNilValue(t *testing.T) {
	f := GenericFlag{Name: "g"}
	if err := f.register(func([]string, flag.Value) {}); err == nil {
		t.Error("expected error for nil GenericFlag value")
	}
}

func TestGenericFlagRegister(t *testing.T) {
	v := &singleValue{}
	f := GenericFlag{Name: "g", Value: v}
	called := false
	err := f.register(func(names []string, val flag.Value) {
		called = true
		if names[0] != "g" {
			t.Errorf("names = %v", names)
		}
	})
	if err != nil || !called {
		t.Fatalf("register err=%v called=%v", err, called)
	}
}

func TestAllFlagTypesInterface(t *testing.T) {
	// Iterate over all flag types, verifying the interface methods do not panic and basic values are correct.
	flags := []Flag{
		&StringFlag{Name: "s", Value: "v"},
		&IntFlag{Name: "i", Value: 1},
		&Int64Flag{Name: "i64", Value: 2},
		&UintFlag{Name: "u", Value: 3},
		&Float64Flag{Name: "f", Value: 4.5},
		&BoolFlag{Name: "b", Value: true},
		&DurationFlag{Name: "d", Value: time.Second},
		&StringSliceFlag{Name: "ss", Value: []string{"x"}},
		&IntSliceFlag{Name: "is", Value: []int{5}},
		&Float64SliceFlag{Name: "fs", Value: []float64{6.5}},
		&GenericFlag{Name: "g", Value: &singleValue{}},
	}
	for _, f := range flags {
		if len(f.Names()) == 0 || f.Names()[0] == "" {
			t.Errorf("%T: empty names", f)
		}
		// isHidden/source/usage/IsRequired should not panic
		_ = f.isHidden()
		_ = f.usage()
		_ = f.source()
		_ = f.IsRequired()
	}
	// Only BoolFlag takes no value
	if flags[5].TakesValue() {
		t.Error("BoolFlag should not take value")
	}
	for i, f := range flags {
		if i == 5 {
			continue
		}
		if !f.TakesValue() {
			t.Errorf("%T should take value", f)
		}
	}
}

func TestFlagIsRequiredFlag(t *testing.T) {
	if !(&StringFlag{Name: "r", Required: true}).IsRequired() {
		t.Error("required StringFlag should report required")
	}
	if (&StringFlag{Name: "o"}).IsRequired() {
		t.Error("optional StringFlag should not be required")
	}
	if !(&GenericFlag{Name: "g", Value: &singleValue{}, Required: true}).IsRequired() {
		t.Error("required GenericFlag should report required")
	}
}

func TestStringFlagSourceAndHidden(t *testing.T) {
	f := StringFlag{Name: "x", Hidden: true, Sources: EnvVars("X")}
	if !f.isHidden() {
		t.Error("should be hidden")
	}
	src := f.source()
	if src == nil {
		t.Fatal("source should not be nil")
	}
	if src.String() != "$X" {
		t.Errorf("source = %q", src.String())
	}
}

func TestSubcommandFlagUsesValueSource(t *testing.T) {
	t.Setenv("CLI_TEST_HOST", "http://es:9200")
	global := []Flag{&StringFlag{Name: "host", Value: "http://127.0.0.1:9200", Sources: EnvVars("CLI_TEST_HOST")}}

	sub := &Command{Name: "run", Flags: append([]Flag{}, global...)}
	sub.Action = func(ctx context.Context, in *Input, out *Output) error {
		if got := in.String("host"); got != "http://es:9200" {
			t.Errorf("subcommand host = %q, want env value", got)
		}
		return nil
	}
	app := &Command{Name: "demo", Flags: append([]Flag{}, global...), Subcommands: []*Command{sub}}
	if _, _, err := runApp(app, "run"); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	subOnly := &Command{Name: "run", Flags: append([]Flag{}, global...)}
	subOnly.Action = func(ctx context.Context, in *Input, out *Output) error {
		if got := in.String("host"); got != "http://es:9200" {
			t.Errorf("subcommand-only host = %q, want env value", got)
		}
		return nil
	}
	appOnly := &Command{Name: "demo", Subcommands: []*Command{subOnly}}
	if _, _, err := runApp(appOnly, "run"); err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestFlagPriorityAcrossCommandTree(t *testing.T) {
	t.Setenv("CLI_TEST_HOST", "http://env:9200")
	source := func() Flag {
		return &StringFlag{Name: "host", Value: "http://default:9200", Sources: EnvVars("CLI_TEST_HOST")}
	}

	tests := []struct {
		name string
		root []Flag
		sub  []Flag
		args []string
		want string
	}{
		{
			name: "root command line beats subcommand source",
			root: []Flag{source()},
			sub:  []Flag{source()},
			args: []string{"--host", "http://cli:9200", "run"},
			want: "http://cli:9200",
		},
		{
			name: "subcommand command line beats root source",
			root: []Flag{source()},
			sub:  []Flag{source()},
			args: []string{"run", "--host", "http://cli:9200"},
			want: "http://cli:9200",
		},
		{
			name: "parent source beats subcommand default",
			root: []Flag{source()},
			sub:  []Flag{&StringFlag{Name: "host", Value: "http://local:9200"}},
			args: []string{"run"},
			want: "http://env:9200",
		},
		{
			name: "subcommand source beats root default",
			root: []Flag{&StringFlag{Name: "host", Value: "http://root:9200"}},
			sub:  []Flag{source()},
			args: []string{"run"},
			want: "http://env:9200",
		},
		{
			name: "default falls back to nearest declaration",
			root: []Flag{&StringFlag{Name: "host", Value: "http://root:9200"}},
			sub:  []Flag{&StringFlag{Name: "host", Value: "http://local:9200"}},
			args: []string{"run"},
			want: "http://local:9200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			sub := &Command{Name: "run", Flags: tt.sub}
			sub.Action = func(ctx context.Context, in *Input, out *Output) error {
				got = in.String("host")
				return nil
			}
			app := &Command{Name: "demo", Flags: tt.root, Subcommands: []*Command{sub}}
			if _, _, err := runApp(app, tt.args...); err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("host = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFlagSourceAliasesMarkedSet(t *testing.T) {
	t.Setenv("CLI_TEST_LANG", "zh")
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{&StringFlag{Name: "lang", Aliases: []string{"l"}, Sources: EnvVars("CLI_TEST_LANG")}}
	var setByName, setByAlias bool
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		setByName = in.IsSet("lang")
		setByAlias = in.IsSet("l")
		return nil
	}
	if _, _, err := runApp(app); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !setByName || !setByAlias {
		t.Errorf("IsSet by name = %v, by alias = %v; want both true", setByName, setByAlias)
	}
}

func TestTypedFlagsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		flag Flag
		args []string
	}{
		{"int", &IntFlag{Name: "n"}, []string{"--n", "abc"}},
		{"int64", &Int64Flag{Name: "n"}, []string{"--n", "abc"}},
		{"uint", &UintFlag{Name: "n"}, []string{"--n", "-1"}},
		{"float64", &Float64Flag{Name: "n"}, []string{"--n", "xyz"}},
		{"duration", &DurationFlag{Name: "n"}, []string{"--n", "soon"}},
		{"int slice", &IntSliceFlag{Name: "n"}, []string{"--n", "1,abc"}},
		{"float slice", &Float64SliceFlag{Name: "n"}, []string{"--n", "1.5,abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			app.Name = "prog"
			app.Flags = []Flag{tt.flag}
			app.Action = func(context.Context, *Input, *Output) error { return nil }
			_, _, err := runApp(app, tt.args...)
			var ue *UsageError
			if !errors.As(err, &ue) {
				t.Fatalf("expected UsageError, got %v", err)
			}
		})
	}
}

func TestTypedFlagsAcceptValidValues(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{
		&IntFlag{Name: "count"},
		&Float64Flag{Name: "ratio"},
		&DurationFlag{Name: "timeout"},
		&IntSliceFlag{Name: "ports"},
	}
	var count, ports int
	var ratio float64
	var timeout time.Duration
	app.Action = func(ctx context.Context, in *Input, out *Output) error {
		count = in.Int("count")
		ratio = in.Float64("ratio")
		timeout = in.Duration("timeout")
		ports = len(in.IntSlice("ports"))
		return nil
	}
	if _, _, err := runApp(app, "--count", "3", "--ratio", "1.5", "--timeout", "2s", "--ports", "80,443"); err != nil {
		t.Fatalf("valid values rejected: %v", err)
	}
	if count != 3 || ratio != 1.5 || timeout != 2*time.Second || ports != 2 {
		t.Errorf("parsed values = %d, %v, %v, %d", count, ratio, timeout, ports)
	}
}

func TestValueSourceValidatesTypedValues(t *testing.T) {
	t.Setenv("CLI_TEST_COUNT", "nope")
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{&IntFlag{Name: "count", Sources: EnvVars("CLI_TEST_COUNT")}}
	app.Action = func(context.Context, *Input, *Output) error { return nil }
	_, _, err := runApp(app)
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UsageError from invalid env value, got %v", err)
	}
}

func TestRequiredFlagSatisfiedByValueSourceOnSubcommand(t *testing.T) {
	t.Setenv("CLI_TEST_TOKEN", "from-env")
	app := New()
	app.Name = "prog"
	app.Subcommands = []*Command{{
		Name:  "run",
		Flags: []Flag{&StringFlag{Name: "token", Required: true, Sources: EnvVars("CLI_TEST_TOKEN")}},
		Action: func(ctx context.Context, in *Input, out *Output) error {
			if got := in.String("token"); got != "from-env" {
				t.Errorf("token = %q", got)
			}
			return nil
		},
	}}
	if _, _, err := runApp(app, "run"); err != nil {
		t.Fatalf("required flag should be satisfied by its source: %v", err)
	}
}

func TestLocalFlagShadowsInheritedGlobalFlag(t *testing.T) {
	app := New()
	app.Name = "prog"
	app.Subcommands = []*Command{{
		Name:  "run",
		Flags: []Flag{&StringFlag{Name: "quiet", Usage: "local quiet flag"}},
		Action: func(ctx context.Context, in *Input, out *Output) error {
			out.Info("still visible")
			if got := in.String("quiet"); got != "local" {
				t.Errorf("quiet = %q, want local flag value", got)
			}
			return nil
		},
	}}
	stdout, _, err := runApp(app, "run", "--quiet", "local")
	if err != nil {
		t.Fatalf("local flag should shadow the inherited global: %v", err)
	}
	if !strings.Contains(stdout, "still visible") {
		t.Errorf("local --quiet should not change verbosity:\n%s", stdout)
	}
}

func TestValueSourceInvalidBoolRejected(t *testing.T) {
	t.Setenv("CLI_TEST_FORCE", "yes")
	app := New()
	app.Name = "prog"
	app.Flags = []Flag{&BoolFlag{Name: "force", Sources: EnvVars("CLI_TEST_FORCE")}}
	app.Action = func(context.Context, *Input, *Output) error { return nil }
	_, _, err := runApp(app)
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UsageError from invalid bool env value, got %v", err)
	}
}
