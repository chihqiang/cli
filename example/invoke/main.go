// invoke —— demos calling command B from inside command A.
//
// Typical scenario: the deploy command needs to run build first, then test.
//
// Run:
//
//	go run ./example/invoke deploy --env prod          → shared functions (recommended)
//	go run ./example/invoke deploy --env prod --skip-test
//	go run ./example/invoke deploy --env prod --count 3
//	go run ./example/invoke chain                       → invoke subcommand Run directly
//	go run ./example/invoke dispatch                    → re-dispatch through the root command
//
// Comparing the three approaches:
//  1. Shared functions (recommended): extract the logic into standalone functions that both A and B reuse.
//     No side effects, no state pollution, clearest and easiest to test.
//  2. Direct Run invocation: use in.FindCommand("build") to locate the subcommand and call its Run,
//     which performs a full flag parse; note that B's After hook runs twice (execute + finish).
//  3. Re-dispatch through the root command: hand "b args..." back to the root App for dispatch again,
//     the simplest, but the root's Before/After hooks run again.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

// buildAction is the core logic of the build command; deploy reuses it via this function.
func buildAction(ctx context.Context, in *cli.Input, out *cli.Output) error {
	out.Infof("building %s ...", in.String("env"))
	return nil
}

// testAction is the core logic of the test command; deploy reuses it via this function.
func testAction(ctx context.Context, in *cli.Input, out *cli.Output) error {
	out.Infof("running tests (%d) ...", in.Int("count"))
	return nil
}

func main() {
	app := &cli.App{
		Name:  "invoke",
		Usage: "demo calling one command from another",
		Subcommands: []*cli.Command{
			// The command being called: build
			{
				Name:  "build",
				Usage: "build the project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "env", Value: "dev", Usage: "target environment"},
				},
				Action: buildAction,
			},
			// The command being called: test
			{
				Name:  "test",
				Usage: "run the tests",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "count", Value: 1, Usage: "how many times"},
				},
				Action: testAction,
			},
			// Approach 1 (recommended): deploy reuses the shared functions of build/test
			{
				Name:  "deploy",
				Usage: "build + test then deploy (shared functions)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "env", Value: "dev", Usage: "target environment"},
					&cli.BoolFlag{Name: "skip-test", Usage: "skip tests"},
					// Reusing testAction requires the count flag it depends on;
					// otherwise in.Int("count") returns 0 in the deploy context.
					&cli.IntFlag{Name: "count", Value: 1, Usage: "how many times"},
				},
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					// Reuse build's core logic (equivalent to running the build command)
					if err := buildAction(ctx, in, out); err != nil {
						return err
					}
					if !in.Bool("skip-test") {
						// Reuse test's core logic (equivalent to running the test command)
						if err := testAction(ctx, in, out); err != nil {
							return err
						}
					}
					out.Successf("deployed to %s", in.String("env"))
					return nil
				},
			},
			// Approach 2: invoke the subcommand's Run directly (full flag parsing)
			{
				Name:  "chain",
				Usage: "call subcommand via its Run method",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					// in.FindCommand is provided by the library; it recursively finds a subcommand by name/alias
					build := in.FindCommand("build")
					if build == nil {
						return &cli.NotFoundError{Command: "build"}
					}
					// Note: if build defines an After hook, it will run twice
					return build.Run(ctx, []string{"build", "--env", "prod"})
				},
			},
			// Approach 3: hand the args back to the root command for re-dispatch
			{
				Name:  "dispatch",
				Usage: "re-dispatch args through the root app",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					// Note: the root command's Before/After hooks run again
					return in.Root().Run(ctx, []string{"invoke", "build", "--env", "prod"})
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
