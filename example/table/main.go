// table —— demos table rendering: multiple styles, section titles, and separators.
//
// Run: go run ./example/table default
//
//	go run ./example/table box
//	go run ./example/table markdown
//	go run ./example/table borderless
//	go run ./example/table advanced
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/cli"
)

func main() {
	app := &cli.App{
		Name:  "table",
		Usage: "demo table rendering",
		Subcommands: []*cli.Command{
			{
				Name:  "default",
				Usage: "default ascii table",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					out.Table([]string{"Name", "Score"}, [][]string{
						{"Alice", "90"}, {"Bob", "85"}, {"Carol", "95"},
					})
					return nil
				},
			},
			{
				Name:  "box",
				Usage: "unicode box table",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					t := cli.NewTable()
					t.SetStyle("box")
					t.SetHeader([]string{"Name", "Score"}, cli.AlignLeft)
					t.SetRows([][]string{{"Alice", "90"}, {"Bob", "85"}, {"Carol", "95"}}, cli.AlignLeft)
					out.RenderTable(t)
					return nil
				},
			},
			{
				Name:  "markdown",
				Usage: "markdown table",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					t := cli.NewTable()
					t.SetStyle("markdown")
					t.SetHeader([]string{"Name", "Score"}, cli.AlignLeft)
					t.SetRows([][]string{{"Alice", "90"}, {"Bob", "85"}, {"Carol", "95"}}, cli.AlignLeft)
					out.RenderTable(t)
					return nil
				},
			},
			{
				Name:  "borderless",
				Usage: "borderless table",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					t := cli.NewTable()
					t.SetStyle("borderless")
					t.SetHeader([]string{"Name", "Score"}, cli.AlignLeft)
					t.SetRows([][]string{{"Alice", "90"}, {"Bob", "85"}, {"Carol", "95"}}, cli.AlignLeft)
					out.RenderTable(t)
					return nil
				},
			},
			{
				Name:  "advanced",
				Usage: "sections, separators and mixed cell types",
				Action: func(ctx context.Context, in *cli.Input, out *cli.Output) error {
					t := cli.NewTable()
					t.SetStyle("box")
					t.SetHeader([]string{"Item", "Qty", "Price"}, cli.AlignLeft)
					t.AddSection("Fruits")
					t.AddRowAny([]interface{}{"Apple", 3, 1.5}, false)
					t.AddRowAny([]interface{}{"Banana", 2, 0.8}, false)
					t.AddSeparator()
					t.AddSection("Drinks")
					t.AddRowAny([]interface{}{"Coffee", 1, 2.0}, false)
					t.AddRowAny([]interface{}{"Tea", 4, 1.2}, false)
					out.RenderTable(t)
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
