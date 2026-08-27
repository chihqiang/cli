// hello —— minimal intro: one App, one subcommand, one positional argument, one flag.
//
// Run: go run ./example/hello greet            → Hello, World!
//
//	go run ./example/hello greet Alice --shout → HELLO, ALICE!
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

func main() {
	app := &cli.App{
		Name:        "hello",
		Usage:       "say hello to someone",
		Version:     "1.0.0",
		Description: "The minimal console app: one command, one argument, one flag.",
		Subcommands: []*cli.Command{
			{
				Name:    "greet",
				Aliases: []string{"g"},
				Usage:   "greet someone",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "shout", Aliases: []string{"s"}, Usage: "shout the greeting"},
				},
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					name := "World"
					if in.Args().Present() {
						name = in.Args().First()
					}
					if in.Bool("shout") {
						out.Infof("HELLO, %s!", name)
						return nil
					}
					out.Infof("Hello, %s!", name)
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
