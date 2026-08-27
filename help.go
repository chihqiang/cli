package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// showHelp displays command help (App help for the root command, command help for subcommands).
func (c *Command) showHelp(ctx context.Context) error {
	out := c.output()
	if c.app == c {
		return c.renderAppHelp(out)
	}
	return c.renderCommandHelp(out)
}

// renderAppHelp renders the root command help.
func (c *Command) renderAppHelp(out *Output) error {
	prog := programName()

	out.rawf("NAME:")
	out.rawf("   %s - %s", prog, c.Usage)

	out.rawf("")
	out.rawf("USAGE:")
	if len(c.Subcommands) > 0 {
		out.rawf("   %s [global options] command [command options] [arguments...]", prog)
	} else {
		out.rawf("   %s [global options] [arguments...]", prog)
	}

	if c.Version != "" {
		out.rawf("")
		out.rawf("VERSION:")
		out.rawf("   %s", c.Version)
	}

	if c.Description != "" {
		out.rawf("")
		out.rawf("DESCRIPTION:")
		out.rawf("   %s", c.Description)
	}

	if len(c.Subcommands) > 0 {
		out.rawf("")
		out.rawf("COMMANDS:")
		cmds := make([]*Command, 0, len(c.Subcommands))
		cmds = append(cmds, c.Subcommands...)
		// Builtin help/version
		hasHelp := false
		hasVersion := false
		for _, sub := range cmds {
			if sub.Name == "help" {
				hasHelp = true
			}
			if sub.Name == "version" {
				hasVersion = true
			}
		}
		if !hasHelp {
			cmds = append(cmds, &Command{Name: "help", Usage: "Shows a list of commands or help for one command"})
		}
		if c.Version != "" && !hasVersion {
			cmds = append(cmds, &Command{Name: "version", Usage: "Print the version"})
		}
		sort.SliceStable(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

		width := 0
		for _, sub := range cmds {
			name := sub.Name
			if len(sub.Aliases) > 0 {
				name += ", " + strings.Join(sub.Aliases, ", ")
			}
			if stringWidth(name) > width {
				width = stringWidth(name)
			}
		}
		for _, sub := range cmds {
			if sub.Hidden {
				continue
			}
			name := sub.Name
			if len(sub.Aliases) > 0 {
				name += ", " + strings.Join(sub.Aliases, ", ")
			}
			out.rawf("   %-*s  %s", width, name, sub.Usage)
		}
	}

	out.rawf("")
	out.rawf("GLOBAL OPTIONS:")
	c.renderFlags(out, c.Flags, true)

	return nil
}

// renderCommandHelp renders the subcommand help.
func (c *Command) renderCommandHelp(out *Output) error {
	prog := programName()

	usage := c.UsageText
	if usage == "" {
		usage = fmt.Sprintf("%s %s", prog, c.Name)
	}

	out.rawf("NAME:")
	if c.Usage != "" {
		out.rawf("   %s %s - %s", prog, c.Name, c.Usage)
	} else {
		out.rawf("   %s %s", prog, c.Name)
	}

	out.rawf("")
	out.rawf("USAGE:")
	out.rawf("   %s [command options] [arguments...]", usage)

	if c.Description != "" {
		out.rawf("")
		out.rawf("DESCRIPTION:")
		out.rawf("   %s", c.Description)
	}

	if len(c.Flags) > 0 {
		out.rawf("")
		out.rawf("OPTIONS:")
		c.renderFlags(out, c.Flags, false)
	}

	return nil
}

// renderFlags renders the flag listing, including the help/version builtins and global flags.
func (c *Command) renderFlags(out *Output, flags []Flag, global bool) {
	// Aggregate (builtin help + optional version + user flags + global builtins)
	type entry struct {
		names string
		usage string
	}
	entries := make([]entry, 0)

	addBuiltin := func(name string, usage string) {
		entries = append(entries, entry{names: name, usage: usage})
	}

	if global {
		addBuiltin("--help, -h", "show help")
		if c.Version != "" {
			addBuiltin("--version, -v", "print the version")
		}
	}
	for _, f := range flags {
		if f.isHidden() {
			continue
		}
		names := "--" + f.Names()[0]
		for _, a := range f.Names()[1:] {
			if len(a) == 1 {
				names += ", -" + a
			} else {
				names += ", --" + a
			}
		}
		if f.TakesValue() {
			names += " value"
		}
		entries = append(entries, entry{names: names, usage: f.usage()})
	}
	if global {
		for _, f := range c.builtinFlags {
			if f.isHidden() {
				continue
			}
			entries = append(entries, entry{names: "--" + f.Names()[0], usage: f.usage()})
		}
	}

	width := 0
	for _, e := range entries {
		if stringWidth(e.names) > width {
			width = stringWidth(e.names)
		}
	}
	for _, e := range entries {
		out.rawf("   %-*s  %s", width, e.names, e.usage)
	}
}
