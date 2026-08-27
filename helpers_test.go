package cli

import (
	"bytes"
	"context"
)

// newBufOutput returns an undecorated Output that captures stdout/stderr.
func newBufOutput() (*Output, *bytes.Buffer, *bytes.Buffer) {
	o := NewOutput()
	o.formatter.SetDecorated(false)
	var out, errb bytes.Buffer
	o.stdout = &out
	o.stderr = &errb
	return o, &out, &errb
}

// captureApp switches the app's shared output to a capturable buffer, returning the output and two buffers.
func captureApp(app *Command) (*Output, *bytes.Buffer, *bytes.Buffer) {
	o, out, errb := newBufOutput()
	app.out = o
	return o, out, errb
}

// runApp runs the app and captures its stdout/stderr text.
func runApp(app *Command, args ...string) (stdout, stderr string, err error) {
	_, out, errb := captureApp(app)
	full := append([]string{"prog"}, args...)
	err = app.Run(context.Background(), full)
	return out.String(), errb.String(), err
}
