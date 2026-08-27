// flags —— demos every flag type: string/int/float/bool/duration/slice/required/env.
//
// Run: go run ./example/flags --token abc --count 3 --ports 80,443 --tags a --tags b
//
//	DEMO_ENV=fallback go run ./example/flags --token abc
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

func main() {
	app := &cli.App{
		Name:  "flags",
		Usage: "demo every flag type",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "Go", Usage: "a string flag"},
			&cli.IntFlag{Name: "count", Aliases: []string{"c"}, Value: 1, Usage: "an int flag"},
			&cli.Int64Flag{Name: "big", Value: 0, Usage: "an int64 flag"},
			&cli.UintFlag{Name: "small", Value: 0, Usage: "a uint flag"},
			&cli.Float64Flag{Name: "ratio", Value: 0.5, Usage: "a float64 flag"},
			&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "a bool flag"},
			&cli.DurationFlag{Name: "timeout", Value: 0, Usage: "a duration flag"},
			&cli.StringSliceFlag{Name: "tags", Value: []string{"default"}, Usage: "a string slice flag"},
			&cli.IntSliceFlag{Name: "ports", Usage: "an int slice flag"},
			&cli.StringFlag{Name: "token", Required: true, Usage: "a required flag"},
			&cli.StringFlag{Name: "env", Sources: cli.EnvVars("DEMO_ENV"), Usage: "a flag with env fallback"},
		},
		Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
			out.Table([]string{"Flag", "Value"}, [][]string{
				{"--name", in.String("name")},
				{"--count", fmt.Sprint(in.Int("count"))},
				{"--big", fmt.Sprint(in.Int64("big"))},
				{"--small", fmt.Sprint(in.Uint("small"))},
				{"--ratio", fmt.Sprint(in.Float64("ratio"))},
				{"--force", fmt.Sprint(in.Bool("force"))},
				{"--timeout", in.Duration("timeout").String()},
				{"--tags", fmt.Sprint(in.StringSlice("tags"))},
				{"--ports", fmt.Sprint(in.IntSlice("ports"))},
				{"--token", in.String("token")},
				{"--env", in.String("env")},
			})
			out.Infof("--force explicitly set? %v", in.IsSet("force"))
			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
