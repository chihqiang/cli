package cli

import (
	"io"
	"strconv"
	"strings"
	"time"
)

// Args holds the positional arguments of a command.
type Args []string

// First returns the first argument (empty string when none exist).
func (a Args) First() string {
	if len(a) > 0 {
		return a[0]
	}
	return ""
}

// Get returns the i-th argument.
func (a Args) Get(i int) string { return a[i] }

// Slice returns all arguments.
func (a Args) Slice() []string { return []string(a) }

// Len returns the number of arguments.
func (a Args) Len() int { return len(a) }

// Present reports whether there are any arguments.
func (a Args) Present() bool { return len(a) != 0 }

// Tail returns the remaining arguments after the first.
func (a Args) Tail() []string {
	if len(a) > 1 {
		return []string(a)[1:]
	}
	return nil
}

// Input carries the input side during command execution: the current command, positional
// arguments, and flag values. It is separated from Output to avoid merging input and
// output into a single object.
type Input struct {
	cmd   *Command
	args  Args
	stdin io.Reader
}

// newInput constructs an input object for the command.
func (c *Command) newInput(args []string) *Input {
	in := &Input{cmd: c, args: args}
	if c.app != nil {
		in.stdin = c.app.Stdin
	}
	return in
}

// Command returns the currently executing command.
func (in *Input) Command() *Command { return in.cmd }

// Root returns the root command (App).
func (in *Input) Root() *Command { return in.cmd.app }

// FindCommand recursively finds a subcommand under the root command (App) by name or
// alias; returns nil when not found.
//
// Often used to invoke command B from inside the Action of command A, for example:
//
//	if b := in.FindCommand("build"); b != nil {
//	    return b.Run(ctx, []string{"build", "--env", "prod"})
//	}
func (in *Input) FindCommand(name string) *Command {
	if in.cmd == nil || in.cmd.app == nil {
		return nil
	}
	return in.cmd.app.findSubcommandDeep(name)
}

// Args returns the current command's positional arguments.
func (in *Input) Args() Args { return in.args }

// NArg returns the number of positional arguments.
func (in *Input) NArg() int { return len(in.args) }

// lookup finds a flag's underlying value by walking up from the current command
// (along the parent chain).
func (in *Input) lookup(name string) (interface{}, bool) {
	for cmd := in.cmd; cmd != nil; cmd = cmd.parent {
		if v, ok := cmd.values[name]; ok {
			return v, true
		}
	}
	return nil, false
}

// String returns the string value of a flag.
func (in *Input) String(name string) string {
	v, ok := in.lookup(name)
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case *singleValue:
		return t.val
	case *sliceValue:
		if len(t.vals) > 0 {
			return t.vals[len(t.vals)-1]
		}
	case *boolValue:
		return strconv.FormatBool(t.val)
	}
	return ""
}

// Bool returns the value of a bool flag.
func (in *Input) Bool(name string) bool {
	v, ok := in.lookup(name)
	if !ok {
		return false
	}
	if bv, ok := v.(*boolValue); ok {
		return bv.val
	}
	return false
}

// Int returns the value of an integer flag.
func (in *Input) Int(name string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(in.String(name)))
	return n
}

// Int64 returns the value of a 64-bit integer flag.
func (in *Input) Int64(name string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(in.String(name)), 10, 64)
	return n
}

// Uint returns the value of an unsigned integer flag.
func (in *Input) Uint(name string) uint {
	n, _ := strconv.ParseUint(strings.TrimSpace(in.String(name)), 10, 64)
	return uint(n)
}

// Float64 returns the value of a floating-point flag.
func (in *Input) Float64(name string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(in.String(name)), 64)
	return f
}

// Duration returns the value of a duration flag.
func (in *Input) Duration(name string) time.Duration {
	d, _ := time.ParseDuration(in.String(name))
	return d
}

// StringSlice returns the value of a string slice flag.
func (in *Input) StringSlice(name string) []string {
	v, ok := in.lookup(name)
	if !ok {
		return nil
	}
	if sv, ok := v.(*sliceValue); ok {
		return append([]string{}, sv.vals...)
	}
	if sv, ok := v.(*singleValue); ok {
		return []string{sv.val}
	}
	return nil
}

// IntSlice returns the value of an integer slice flag.
func (in *Input) IntSlice(name string) []int {
	vals := in.StringSlice(name)
	out := make([]int, 0, len(vals))
	for _, s := range vals {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// Float64Slice returns the value of a floating-point slice flag.
func (in *Input) Float64Slice(name string) []float64 {
	vals := in.StringSlice(name)
	out := make([]float64, 0, len(vals))
	for _, s := range vals {
		if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			out = append(out, f)
		}
	}
	return out
}

// Generic returns the underlying value of a generic flag.
func (in *Input) Generic(name string) interface{} {
	v, _ := in.lookup(name)
	return v
}

// IsSet reports whether the flag was explicitly set (command line or environment variable).
func (in *Input) IsSet(name string) bool {
	for cmd := in.cmd; cmd != nil; cmd = cmd.parent {
		if cmd.setFlags[name] {
			return true
		}
	}
	return false
}

// IsInteractive reports whether interactive Q&A is allowed; it is false with
// --no-interaction or when stdin is not a terminal.
func (in *Input) IsInteractive() bool {
	if in.Bool("no-interaction") {
		return false
	}
	root := in.cmd.app
	if root != nil {
		return root.interactive
	}
	return true
}
