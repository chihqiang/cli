package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Output verbosity levels.
const (
	VerbosityQuiet  = 0 // Keep only warnings and errors
	VerbosityNormal = 1
	VerbosityDebug  = 4 // --verbose
)

// Output handles common log output: Info / Success / Warn / Error / Debug / Verbose,
// plus table rendering and interactive Q&A. Color is enabled automatically on terminals.
type Output struct {
	verbosity int
	formatter *Formatter
	stdout    io.Writer
	stderr    io.Writer
	root      *Command      // Root command reference, for Q&A interactivity checks and injectable stdin
	bufReader *bufio.Reader // Persistent buffered reader for Q&A (shared across consecutive questions to avoid losing data)
}

// NewOutput creates an output object writing to stdout/stderr. Color detection is delegated to fatih/color.
func NewOutput() *Output {
	f := NewFormatter()
	f.SetDecorated(!color.NoColor)
	return &Output{
		verbosity: VerbosityNormal,
		formatter: f,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
	}
}

// SetVerbosity sets the output level.
func (o *Output) SetVerbosity(level int) { o.verbosity = level }

// GetVerbosity returns the output level.
func (o *Output) GetVerbosity() int { return o.verbosity }

// Info prints an informational message.
func (o *Output) Info(args ...interface{}) {
	if o.verbosity == VerbosityQuiet {
		return
	}
	o.emit("info", fmt.Sprint(args...), o.stdout)
}

// Infof prints a formatted informational message.
func (o *Output) Infof(format string, args ...interface{}) {
	if o.verbosity == VerbosityQuiet {
		return
	}
	o.emit("info", fmt.Sprintf(format, args...), o.stdout)
}

// Success prints a success message.
func (o *Output) Success(args ...interface{}) {
	if o.verbosity == VerbosityQuiet {
		return
	}
	o.emit("success", fmt.Sprint(args...), o.stdout)
}

// Successf prints a formatted success message.
func (o *Output) Successf(format string, args ...interface{}) {
	if o.verbosity == VerbosityQuiet {
		return
	}
	o.emit("success", fmt.Sprintf(format, args...), o.stdout)
}

// Warn prints a warning message (not suppressed by --quiet).
func (o *Output) Warn(args ...interface{}) {
	o.emit("warning", fmt.Sprint(args...), o.stderr)
}

// Warnf prints a formatted warning message.
func (o *Output) Warnf(format string, args ...interface{}) {
	o.emit("warning", fmt.Sprintf(format, args...), o.stderr)
}

// Error prints an error message (written to stderr, not suppressed by --quiet).
func (o *Output) Error(args ...interface{}) {
	o.emit("error", fmt.Sprint(args...), o.stderr)
}

// Errorf prints a formatted error message.
func (o *Output) Errorf(format string, args ...interface{}) {
	o.emit("error", fmt.Sprintf(format, args...), o.stderr)
}

// Debug prints a debug message (requires --verbose).
func (o *Output) Debug(args ...interface{}) {
	if o.verbosity < VerbosityDebug {
		return
	}
	o.emit("comment", fmt.Sprint(args...), o.stdout)
}

// Debugf prints a formatted debug message.
func (o *Output) Debugf(format string, args ...interface{}) {
	if o.verbosity < VerbosityDebug {
		return
	}
	o.emit("comment", fmt.Sprintf(format, args...), o.stdout)
}

// Verbose prints a verbose message (requires --verbose).
func (o *Output) Verbose(args ...interface{}) {
	if o.verbosity < VerbosityDebug {
		return
	}
	o.emit("comment", fmt.Sprint(args...), o.stdout)
}

// Verbosef prints a formatted verbose message.
func (o *Output) Verbosef(format string, args ...interface{}) {
	if o.verbosity < VerbosityDebug {
		return
	}
	o.emit("comment", fmt.Sprintf(format, args...), o.stdout)
}

// raw prints a line as-is (no style wrapping; used for help text etc.).
func (o *Output) raw(msg string) {
	fmt.Fprintln(o.stdout, msg)
}

// rawf prints a formatted line as-is.
func (o *Output) rawf(format string, args ...interface{}) {
	fmt.Fprintf(o.stdout, format+"\n", args...)
}

// emit prints one line with the given style.
func (o *Output) emit(style, msg string, w io.Writer) {
	if o.formatter.IsDecorated() {
		msg = o.formatter.Format("<" + style + ">" + msg + "</" + style + ">")
	}
	fmt.Fprintln(w, msg)
}

// ------------------------- Tables -------------------------

// RenderTable renders and prints a custom Table (style/sections/separators supported)
// and returns the text.
func (o *Output) RenderTable(t *Table) string {
	content := t.Render()
	o.raw(content)
	return content
}

// Table renders and prints a table, returning the rendered text.
func (o *Output) Table(header []string, rows [][]string) string {
	t := NewTable()
	t.SetHeader(header, AlignLeft)
	t.SetRows(rows, AlignLeft)
	return o.RenderTable(t)
}

// TableAny renders and prints a table with arbitrary cell types (cells are stringified automatically).
func (o *Output) TableAny(header []string, rows [][]interface{}) string {
	t := NewTable()
	t.SetHeader(header, AlignLeft)
	cells := make([][]string, len(rows))
	for i, r := range rows {
		cells[i] = stringifyRow(r)
	}
	t.SetRows(cells, AlignLeft)
	return o.RenderTable(t)
}

// ------------------------- Questions -------------------------

// stdin returns the injectable standard input (defaults to os.Stdin).
func (o *Output) stdin() io.Reader {
	if o.root != nil && o.root.Stdin != nil {
		return o.root.Stdin
	}
	return os.Stdin
}

// bufioReader returns the shared persistent buffered reader, reused between consecutive questions.
func (o *Output) bufioReader() *bufio.Reader {
	if o.bufReader == nil {
		o.bufReader = bufio.NewReader(o.stdin())
	}
	return o.bufReader
}

// isInteractive reports whether interactive Q&A is allowed; it is false with
// --no-interaction or when stdin is not a terminal.
func (o *Output) isInteractive() bool {
	if o.root != nil {
		if bv, ok := o.root.values["no-interaction"].(*boolValue); ok && bv.val {
			return false
		}
		return o.root.interactive
	}
	return true
}

// Ask asks a free-form question and returns the validated answer.
func (o *Output) Ask(question string, defaultVal interface{}) (interface{}, error) {
	q := NewQuestion(question, defaultVal)
	if !o.isInteractive() {
		return q.GetDefault(), nil
	}
	return NewAsk(o, q).Run()
}

// AskString asks a string question.
func (o *Output) AskString(question, defaultVal string) (string, error) {
	q := NewQuestion(question, defaultVal)
	if !o.isInteractive() {
		return q.GetDefault().(string), nil
	}
	ans, err := NewAsk(o, q).Run()
	if err != nil {
		return "", err
	}
	return ans.(string), nil
}

// AskHidden asks a hidden-input question (no echo).
func (o *Output) AskHidden(question string) (string, error) {
	q := NewQuestion(question, nil)
	q.SetHidden(true)
	if !o.isInteractive() {
		return "", nil
	}
	ans, err := NewAsk(o, q).Run()
	if err != nil {
		return "", err
	}
	return ans.(string), nil
}

// Confirm asks a yes/no question.
func (o *Output) Confirm(question string, defaultVal bool) (bool, error) {
	q := NewConfirmation(question, defaultVal, "")
	if !o.isInteractive() {
		return defaultVal, nil
	}
	ans, err := NewAsk(o, q).Run()
	if err != nil {
		return false, err
	}
	return ans == "yes", nil
}

// Choice asks a multiple-choice question; choices is a key->label mapping.
func (o *Output) Choice(question string, choices map[string]string, defaultVal interface{}) (interface{}, error) {
	q := NewChoice(question, choices, defaultVal)
	if !o.isInteractive() {
		return q.GetDefault(), nil
	}
	return NewAsk(o, q).Run()
}
