// context —— demos passing values through context.Context across the command tree.
//
// The root (global) Before hook defines values on the context; they are visible to
// every subcommand's Action/After. A subcommand's own Before can further derive the
// context for its own Action/After.
//
// Note: global flags must come before the subcommand name (flag parsing stops at the
// first positional argument when subcommands are declared).
//
// Run:
//
//	go run ./example/context --user alice run
//	go run ./example/context nested --project web
//	go run ./example/context run              → user falls back to "anonymous"
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

// Typed, unexported context keys avoid accidental collisions.
type ctxKey int

const (
	keyUser ctxKey = iota
	keyTrace
	keyProject
)

// globalBefore injects a user identity and a trace ID into the context. These values
// are available to every subcommand via ctx.Value.
func globalBefore(ctx context.Context, in *cli.Input, out *cli.Output) (context.Context, error) {
	user := in.String("user")
	if user == "" {
		user = "anonymous"
	}
	ctx = context.WithValue(ctx, keyUser, user)
	ctx = context.WithValue(ctx, keyTrace, "trace-12345")
	return ctx, nil
}

func main() {
	app := &cli.App{
		Name:  "context",
		Usage: "demo context passing across the command tree",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "user", Value: "", Usage: "current user (injected into context)"},
		},
		Before: globalBefore,
		Subcommands: []*cli.Command{
			{
				Name:  "run",
				Usage: "read values injected by the global Before hook",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					// Values defined in the root Before hook are visible here.
					user, _ := ctx.Value(keyUser).(string)
					trace, _ := ctx.Value(keyTrace).(string)
					out.Successf("running as %q (trace=%s)", user, trace)
					return nil
				},
			},
			{
				Name:  "nested",
				Usage: "derive the context in a subcommand Before and read it in its Action",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "project", Value: "web", Usage: "project name"},
				},
				Before: func(ctx context.Context, in *cli.Input, out *cli.Output) (context.Context, error) {
					// The subcommand's Before can read the global values and add more.
					user, _ := ctx.Value(keyUser).(string)
					out.Infof("subcommand Before: user from global hook = %s", user)
					return context.WithValue(ctx, keyProject, in.String("project")), nil
				},
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					// Both the global values and the subcommand-derived value are available.
					user, _ := ctx.Value(keyUser).(string)
					trace, _ := ctx.Value(keyTrace).(string)
					project, _ := ctx.Value(keyProject).(string)
					out.Successf("user=%s trace=%s project=%s", user, trace, project)
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
