// commands —— demos command organization & the help system: grouping, hidden commands,
// custom usage, and not-found handling.
//
// Run: go run ./example/commands               → root help (commands grouped)
//
//	go run ./example/commands --help
//	go run ./example/commands add --help     → single-command help
//	go run ./example/commands help add
//	go run ./example/commands missing        → triggers CommandNotFound
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

func main() {
	app := &cli.App{
		Name:        "commands",
		Usage:       "demo command organization & help",
		Version:     "2.0.0",
		Description: "Commands grouped for organization; hidden commands stay invokable but are omitted from listings.",
		// Custom version printing: print an extra line in this example.
		VersionPrinter: func(ctx context.Context, in *cli.Input, out *cli.Output) {
			out.Infof("commands version: %s (custom printer)", in.Command().Version)
		},
		CommandNotFound: func(ctx context.Context, in *cli.Input, out *cli.Output, cmd string) {
			out.Errorf("command %q not found, try `commands`", cmd)
		},
		Subcommands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "add an item",
				UsageText: "commands add [--force] <name>",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force add"},
				},
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Successf("added %q (force=%v)", in.Args().First(), in.Bool("force"))
					return nil
				},
			},
			{
				Name:  "remove",
				Usage: "remove an item",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Infof("removed %q", in.Args().First())
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list items",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Table([]string{"ID", "Name"}, [][]string{{"1", "apple"}, {"2", "banana"}})
					return nil
				},
			},
			{
				Name:   "secret",
				Usage:  "hidden utility command",
				Hidden: true,
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Warn("you found the hidden command")
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
