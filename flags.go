package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ValueSource is the interface for flag value sources.
//
// A flag may declare a value source (Sources); at runtime the first hit wins by priority:
// command line > Sources > default value (Value). The builtin implementation is EnvVars;
// custom implementations are also supported.
type ValueSource interface {
	// Lookup returns the value provided by this source; returns empty string and false when not found.
	Lookup() (string, bool)
	// String returns a human-readable description of the source, used in help text.
	String() string
}

// EnvVars builds a ValueSource that reads from environment variables; it is one implementation of ValueSource.
//
// Usage: Sources: cli.EnvVars("APP_LANG", "LANG")
func EnvVars(envs ...string) ValueSource {
	return envValueSource(append([]string(nil), envs...))
}

// envValueSource reads the first set value from a list of environment variables.
type envValueSource []string

func (e envValueSource) Lookup() (string, bool) {
	for _, env := range e {
		if v, ok := os.LookupEnv(env); ok {
			return v, true
		}
	}
	return "", false
}

func (e envValueSource) String() string {
	if len(e) == 0 {
		return ""
	}
	return "$" + strings.Join([]string(e), ", $")
}

// Flag is the interface for all flag types.
type Flag interface {
	// Names returns the primary name plus aliases.
	Names() []string
	// TakesValue reports whether the flag takes a value (bool flags return false).
	TakesValue() bool
	// IsRequired reports whether the flag is required.
	IsRequired() bool
	// source returns the flag's value source (ValueSource).
	source() ValueSource
	// isHidden reports whether the flag is hidden in help.
	isHidden() bool
	// usage returns the usage text, including default value and value source.
	usage() string
	// register registers the flag's underlying value with the command.
	register(func([]string, flag.Value)) error
}

// flagBase carries the shared logic for all flag types (unexported).
type flagBase struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
}

func (b flagBase) names() []string {
	out := []string{b.Name}
	return append(out, b.Aliases...)
}

func (b flagBase) source() ValueSource { return b.Sources }
func (b flagBase) isRequired() bool    { return b.Required }
func (b flagBase) isHidden() bool      { return b.Hidden }

// usageText generates the usage text, appending the default value and value source.
func (b flagBase) usageText(val interface{}) string {
	u := b.Usage
	if b.DefaultText != "" {
		u += fmt.Sprintf(" (default: %s)", b.DefaultText)
	} else if val != nil && fmt.Sprint(val) != "" {
		if _, isBool := val.(bool); !isBool {
			u += fmt.Sprintf(" (default: %v)", val)
		}
	}
	if b.Sources != nil {
		if desc := b.Sources.String(); desc != "" {
			u += fmt.Sprintf(" [%s]", desc)
		}
	}
	return u
}

// singleValue holds the string value of a single-value flag (each Set overwrites).
type singleValue struct {
	val string
}

func (v *singleValue) String() string { return v.val }
func (v *singleValue) Set(s string) error {
	v.val = s
	return nil
}

// sliceValue holds the string values of a slice flag (supports repeated occurrences and comma-separated values).
type sliceValue struct {
	vals  []string
	isSet bool
}

func (v *sliceValue) String() string { return strings.Join(v.vals, ",") }
func (v *sliceValue) Set(s string) error {
	if !v.isSet {
		v.vals = nil
		v.isSet = true
	}
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			v.vals = append(v.vals, part)
		}
	}
	return nil
}

// boolValue is the underlying value of a bool flag.
type boolValue struct {
	val   bool
	isSet bool
}

func (v *boolValue) String() string { return strconv.FormatBool(v.val) }
func (v *boolValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	v.val = b
	v.isSet = true
	return nil
}
func (v *boolValue) IsBoolFlag() bool { return true }

// StringFlag is a string flag.
type StringFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       string
}

func (f StringFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f StringFlag) Names() []string     { return f.common().names() }
func (f StringFlag) TakesValue() bool    { return true }
func (f StringFlag) IsRequired() bool    { return f.common().isRequired() }
func (f StringFlag) source() ValueSource { return f.common().source() }
func (f StringFlag) isHidden() bool      { return f.common().isHidden() }
func (f StringFlag) usage() string       { return f.common().usageText(f.Value) }
func (f StringFlag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &singleValue{val: f.Value})
	return nil
}

// IntFlag is an integer flag.
type IntFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       int
}

func (f IntFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f IntFlag) Names() []string     { return f.common().names() }
func (f IntFlag) TakesValue() bool    { return true }
func (f IntFlag) IsRequired() bool    { return f.common().isRequired() }
func (f IntFlag) source() ValueSource { return f.common().source() }
func (f IntFlag) isHidden() bool      { return f.common().isHidden() }
func (f IntFlag) usage() string       { return f.common().usageText(f.Value) }
func (f IntFlag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &singleValue{val: strconv.Itoa(f.Value)})
	return nil
}

// Int64Flag is a 64-bit integer flag.
type Int64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       int64
}

func (f Int64Flag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f Int64Flag) Names() []string     { return f.common().names() }
func (f Int64Flag) TakesValue() bool    { return true }
func (f Int64Flag) IsRequired() bool    { return f.common().isRequired() }
func (f Int64Flag) source() ValueSource { return f.common().source() }
func (f Int64Flag) isHidden() bool      { return f.common().isHidden() }
func (f Int64Flag) usage() string       { return f.common().usageText(f.Value) }
func (f Int64Flag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &singleValue{val: strconv.FormatInt(f.Value, 10)})
	return nil
}

// UintFlag is an unsigned integer flag.
type UintFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       uint
}

func (f UintFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f UintFlag) Names() []string     { return f.common().names() }
func (f UintFlag) TakesValue() bool    { return true }
func (f UintFlag) IsRequired() bool    { return f.common().isRequired() }
func (f UintFlag) source() ValueSource { return f.common().source() }
func (f UintFlag) isHidden() bool      { return f.common().isHidden() }
func (f UintFlag) usage() string       { return f.common().usageText(f.Value) }
func (f UintFlag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &singleValue{val: strconv.FormatUint(uint64(f.Value), 10)})
	return nil
}

// Float64Flag is a floating-point flag.
type Float64Flag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       float64
}

func (f Float64Flag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f Float64Flag) Names() []string     { return f.common().names() }
func (f Float64Flag) TakesValue() bool    { return true }
func (f Float64Flag) IsRequired() bool    { return f.common().isRequired() }
func (f Float64Flag) source() ValueSource { return f.common().source() }
func (f Float64Flag) isHidden() bool      { return f.common().isHidden() }
func (f Float64Flag) usage() string       { return f.common().usageText(f.Value) }
func (f Float64Flag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &singleValue{val: strconv.FormatFloat(f.Value, 'g', -1, 64)})
	return nil
}

// BoolFlag is a boolean switch flag.
type BoolFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       bool
}

func (f BoolFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f BoolFlag) Names() []string     { return f.common().names() }
func (f BoolFlag) TakesValue() bool    { return false }
func (f BoolFlag) IsRequired() bool    { return f.common().isRequired() }
func (f BoolFlag) source() ValueSource { return f.common().source() }
func (f BoolFlag) isHidden() bool      { return f.common().isHidden() }
func (f BoolFlag) usage() string       { return f.common().usageText(f.Value) }
func (f BoolFlag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &boolValue{val: f.Value})
	return nil
}

// DurationFlag is a time duration flag.
type DurationFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       time.Duration
}

func (f DurationFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f DurationFlag) Names() []string     { return f.common().names() }
func (f DurationFlag) TakesValue() bool    { return true }
func (f DurationFlag) IsRequired() bool    { return f.common().isRequired() }
func (f DurationFlag) source() ValueSource { return f.common().source() }
func (f DurationFlag) isHidden() bool      { return f.common().isHidden() }
func (f DurationFlag) usage() string       { return f.common().usageText(f.Value) }
func (f DurationFlag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &singleValue{val: f.Value.String()})
	return nil
}

// StringSliceFlag is a string slice flag (repeatable).
type StringSliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       []string
}

func (f StringSliceFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f StringSliceFlag) Names() []string     { return f.common().names() }
func (f StringSliceFlag) TakesValue() bool    { return true }
func (f StringSliceFlag) IsRequired() bool    { return f.common().isRequired() }
func (f StringSliceFlag) source() ValueSource { return f.common().source() }
func (f StringSliceFlag) isHidden() bool      { return f.common().isHidden() }
func (f StringSliceFlag) usage() string       { return f.common().usageText(f.Value) }
func (f StringSliceFlag) register(reg func([]string, flag.Value)) error {
	reg(f.Names(), &sliceValue{vals: append([]string{}, f.Value...)})
	return nil
}

// IntSliceFlag is an integer slice flag.
type IntSliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       []int
}

func (f IntSliceFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f IntSliceFlag) Names() []string     { return f.common().names() }
func (f IntSliceFlag) TakesValue() bool    { return true }
func (f IntSliceFlag) IsRequired() bool    { return f.common().isRequired() }
func (f IntSliceFlag) source() ValueSource { return f.common().source() }
func (f IntSliceFlag) isHidden() bool      { return f.common().isHidden() }
func (f IntSliceFlag) usage() string       { return f.common().usageText(f.Value) }
func (f IntSliceFlag) register(reg func([]string, flag.Value)) error {
	vals := make([]string, 0, len(f.Value))
	for _, v := range f.Value {
		vals = append(vals, strconv.Itoa(v))
	}
	reg(f.Names(), &sliceValue{vals: vals})
	return nil
}

// Float64SliceFlag is a floating-point slice flag.
type Float64SliceFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       []float64
}

func (f Float64SliceFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f Float64SliceFlag) Names() []string     { return f.common().names() }
func (f Float64SliceFlag) TakesValue() bool    { return true }
func (f Float64SliceFlag) IsRequired() bool    { return f.common().isRequired() }
func (f Float64SliceFlag) source() ValueSource { return f.common().source() }
func (f Float64SliceFlag) isHidden() bool      { return f.common().isHidden() }
func (f Float64SliceFlag) usage() string       { return f.common().usageText(f.Value) }
func (f Float64SliceFlag) register(reg func([]string, flag.Value)) error {
	vals := make([]string, 0, len(f.Value))
	for _, v := range f.Value {
		vals = append(vals, strconv.FormatFloat(v, 'g', -1, 64))
	}
	reg(f.Names(), &sliceValue{vals: vals})
	return nil
}

// GenericFlag is a generic flag that allows a custom flag.Value implementation.
type GenericFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	Required    bool
	Hidden      bool
	DefaultText string
	Sources     ValueSource
	Value       flag.Value
}

func (f GenericFlag) common() flagBase {
	return flagBase{Name: f.Name, Aliases: f.Aliases, Usage: f.Usage, Required: f.Required, Hidden: f.Hidden, DefaultText: f.DefaultText, Sources: f.Sources}
}
func (f GenericFlag) Names() []string     { return f.common().names() }
func (f GenericFlag) TakesValue() bool    { return true }
func (f GenericFlag) IsRequired() bool    { return f.common().isRequired() }
func (f GenericFlag) source() ValueSource { return f.common().source() }
func (f GenericFlag) isHidden() bool      { return f.common().isHidden() }
func (f GenericFlag) usage() string       { return f.common().usageText(nil) }
func (f GenericFlag) register(reg func([]string, flag.Value)) error {
	if f.Value == nil {
		return &UsageError{Msg: "GenericFlag \"" + f.Name + "\" requires a non-nil Value"}
	}
	reg(f.Names(), f.Value)
	return nil
}
