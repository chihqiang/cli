// ask —— demos interactive Q&A: free input, hidden input, confirmation, and choice.
//
// Run: go run ./example/ask survey
// Non-interactive (piped) answers: printf 'Zhang\n25\nsecret\ny\npy\n' | go run ./example/ask survey
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

func main() {
	app := &cli.App{
		Name:  "ask",
		Usage: "demo interactive questions",
		// Explicitly inject Stdin: allow interaction even when stdin is not a terminal (e.g. a pipe),
		// which is handy for automated testing.
		Stdin: os.Stdin,
		Subcommands: []*cli.Command{
			{
				Name:  "survey",
				Usage: "ask a few questions",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					name, err := out.AskString("What's your name?", "world")
					if err != nil {
						return err
					}
					age, err := out.Ask("How old are you?", 18)
					if err != nil {
						return err
					}
					pwd, err := out.AskHidden("Secret password?")
					if err != nil {
						return err
					}
					like, err := out.Confirm("Do you like console?", true)
					if err != nil {
						return err
					}
					lang, err := out.Choice("Pick a language", map[string]string{
						"go": "Go", "py": "Python", "rs": "Rust",
					}, "go")
					if err != nil {
						return err
					}
					out.Successf("name=%s age=%v like=%v lang=%v pwd=%q", name, age, like, lang, pwd)
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
