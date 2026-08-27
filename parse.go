package cli

import (
	"flag"
	"fmt"
	"strings"
)

// flagDef describes a registered flag.
type flagDef struct {
	value  flag.Value
	isBool bool
}

// isBoolFlag reports whether value is a bool switch.
func isBoolFlag(v flag.Value) bool {
	if bf, ok := v.(interface{ IsBoolFlag() bool }); ok {
		return bf.IsBoolFlag()
	}
	return false
}

// parseTokens parses command-line tokens. Semantics:
//   - Supports --name value / --name=value / -n value / -n=value
//   - A bool flag only accepts --flag / --flag=true and does not consume the next token
//   - interspersed=false: stops parsing flags after the first positional argument (used for subcommand dispatch)
//   - interspersed=true: positional arguments and flags can be interleaved (used for Action commands)
//   - Everything after -- is treated as positional arguments
//
// Returns the remaining positional arguments.
func parseTokens(args []string, defs map[string]*flagDef, setFlags map[string]bool, interspersed bool) ([]string, error) {
	var positional []string
	onlyArgs := false
	for i := 0; i < len(args); i++ {
		tok := args[i]
		if onlyArgs {
			positional = append(positional, tok)
			continue
		}
		if tok == "--" {
			onlyArgs = true
			continue
		}
		if tok == "-" || !strings.HasPrefix(tok, "-") {
			positional = append(positional, tok)
			if !interspersed {
				positional = append(positional, args[i+1:]...)
				break
			}
			continue
		}

		name, val, hasEq := splitFlagToken(tok)
		def, ok := defs[name]
		if !ok {
			return nil, &UsageError{Msg: fmt.Sprintf("flag provided but not defined: -%s", name)}
		}

		if !hasEq && !def.isBool {
			// Take the value from the next token (use --name=value for values starting with a dash)
			if i+1 < len(args) {
				next := args[i+1]
				if next == "-" || !strings.HasPrefix(next, "-") {
					val = next
					i++
				} else {
					return nil, &UsageError{Msg: fmt.Sprintf("flag needs an argument: -%s", name)}
				}
			} else {
				return nil, &UsageError{Msg: fmt.Sprintf("flag needs an argument: -%s", name)}
			}
		}

		setFlags[name] = true
		if err := def.value.Set(val); err != nil {
			return nil, &UsageError{Msg: fmt.Sprintf("invalid value %q for flag -%s: %v", val, name, err)}
		}
	}
	return positional, nil
}

// splitFlagToken splits tokens like --name=value / -n=value into a name and a value.
func splitFlagToken(tok string) (name, value string, hasEq bool) {
	if strings.HasPrefix(tok, "--") {
		rest := tok[2:]
		if i := strings.Index(rest, "="); i >= 0 {
			return rest[:i], rest[i+1:], true
		}
		return rest, "true", false
	}
	// Short option
	rest := tok[1:]
	if i := strings.Index(rest, "="); i >= 0 {
		return rest[:i], rest[i+1:], true
	}
	return rest, "true", false
}
