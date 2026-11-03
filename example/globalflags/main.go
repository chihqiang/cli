// globalflags —— demos global vs local flags: the same flags declared on the root and on a
// subcommand, env fallback (Sources), required validation and typed values.
//
// Run:
//
//	DEMO_TOKEN=from-env go run ./example/globalflags run
//	DEMO_TOKEN=from-env go run ./example/globalflags run --host es:9200 --retries 3
//	DEMO_TOKEN=from-env go run ./example/globalflags --host root:9200 run   → root command line wins
//	DEMO_TOKEN=from-env go run ./example/globalflags run --verbose         → global flag after the command
//	go run ./example/globalflags run --retries abc                        → usage error (typed validation)
//	go run ./example/globalflags run --help                               → OPTIONS + GLOBAL OPTIONS
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

// globalFlags are declared on the root and reused on the subcommand, so users can put
// them before or after the command name.
func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Usage:   "server address",
			Value:   "http://127.0.0.1:9200",
			Sources: cli.EnvVars("DEMO_HOST"),
		},
		&cli.StringFlag{
			Name:     "token",
			Usage:    "auth token (required)",
			Required: true,
			Sources:  cli.EnvVars("DEMO_TOKEN"),
		},
	}
}

func main() {
	app := &cli.App{
		Name:  "globalflags",
		Usage: "demo global and local flags",
		Flags: globalFlags(),
		Subcommands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run with shared flags plus a local --retries flag",
				Flags: append(globalFlags(),
					&cli.IntFlag{Name: "retries", Aliases: []string{"r"}, Value: 1, Usage: "retry count"},
				),
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Debug("debug line (visible with --verbose, before or after the command)")
					out.Table(
						[]string{"Flag", "Value", "Set"},
						[][]string{
							{"--host", in.String("host"), fmt.Sprint(in.IsSet("host"))},
							{"--token", "(hidden)", fmt.Sprint(in.IsSet("token"))},
							{"--retries", fmt.Sprint(in.Int("retries")), fmt.Sprint(in.IsSet("retries"))},
						},
					)
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Run 'globalflags --help' for usage.")
		os.Exit(1)
	}
}
