// logging —— demos log output and level control.
//
// Run: go run ./example/logging demo              → base level
//
//	go run ./example/logging --verbose demo    → additionally prints Verbose/Debug
//	go run ./example/logging demo --verbose    → global flags also work after the command
//	go run ./example/logging --quiet demo      → only Warn/Error remain
//	go run ./example/logging demo --quiet      → same, with the global flag after the command
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

func main() {
	app := &cli.App{
		Name:  "logging",
		Usage: "demo log levels",
		Subcommands: []*cli.Command{
			{
				Name:  "demo",
				Usage: "print every log level",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Info("info message")
					out.Infof("info with args: %s", "formatted")
					out.Success("success message")
					out.Warn("warning message")
					out.Error("error message")
					out.Verbose("verbose message (only with --verbose)")
					out.Debug("debug message (only with --verbose)")
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
