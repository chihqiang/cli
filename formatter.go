package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	tagRegex       = regexp.MustCompile(`<(([a-z][a-z0-9_=;-]*)|/([a-z][a-z0-9_=;-]*)?)>`)
	escapeTagRegex = regexp.MustCompile(`([^\\]?)<`)
	tagStripRegex  = regexp.MustCompile(`<[^>]*>`)
)

// Formatter renders colors via <style>...</style> tags, delegating the underlying ANSI
// rendering to fatih/color (Symfony-style tag rendering; colors are not self-implemented).
type Formatter struct {
	decorated bool
	styles    map[string]string // Named style -> SGR parameter string (e.g. "31", "1;32")
	stack     []string          // Style stack for nested tags (SGR parameter strings)
}

// NewFormatter creates a formatter with the default styles.
func NewFormatter() *Formatter {
	f := &Formatter{styles: map[string]string{}}
	// SGR parameter strings for the default styles (matching the original gookit builtin tag codes).
	f.styles["error"] = "97;41"     // fg bright white; bg red
	f.styles["info"] = "0;32"       // green
	f.styles["comment"] = "0;33"    // yellow/brown
	f.styles["warning"] = "0;30;43" // fg black; bg yellow
	f.styles["success"] = "1;32"    // bold green
	f.styles["question"] = "30;46"  // fg black; bg cyan
	f.styles["highlight"] = "31"    // red
	return f
}

// Escape escapes style tags in text.
func (f *Formatter) Escape(text string) string {
	return escapeTagRegex.ReplaceAllString(text, "$1\\<")
}

// SetDecorated enables/disables styling. When disabled, plain text is output (color codes stripped).
func (f *Formatter) SetDecorated(decorated bool) { f.decorated = decorated }

// IsDecorated reports whether styling is enabled.
func (f *Formatter) IsDecorated() bool { return f.decorated }

// SetStyle registers a style by name (code is an SGR parameter string, e.g. "32", "1;31").
func (f *Formatter) SetStyle(name string, code string) {
	f.styles[strings.ToLower(name)] = code
}

// HasStyle reports whether the style exists.
func (f *Formatter) HasStyle(name string) bool {
	_, ok := f.styles[strings.ToLower(name)]
	return ok
}

// GetStyle returns the SGR parameter string of a registered style.
func (f *Formatter) GetStyle(name string) (string, error) {
	name = strings.ToLower(name)
	if !f.HasStyle(name) {
		return "", fmt.Errorf("undefined style: %s", name)
	}
	return f.styles[name], nil
}

// Format processes a message, replacing style tags with ANSI codes (rendered via fatih/color).
func (f *Formatter) Format(message string) string {
	offset := 0
	output := ""

	matches := tagRegex.FindAllStringSubmatchIndex(message, -1)
	for _, match := range matches {
		pos := match[0]
		text := message[match[0]:match[1]]

		if pos != 0 && message[pos-1] == '\\' {
			continue
		}

		output += f.applyCurrentStyle(message[offset:pos])
		offset = match[1]

		var tag string
		open := message[pos+1] != '/'
		if open {
			tag = message[match[2]:match[3]]
		} else if match[6] >= 0 {
			tag = message[match[6]:match[7]]
		} else {
			tag = ""
		}

		if !open && tag == "" {
			f.pop()
		} else if code, ok := f.createCodeFromString(strings.ToLower(tag)); !ok {
			output += f.applyCurrentStyle(text)
		} else if open {
			f.push(code)
		} else {
			f.popTo(code)
		}
	}

	output += f.applyCurrentStyle(message[offset:])
	return strings.ReplaceAll(output, "\\<", "<")
}

// fgColorNames / bgColorNames / opCodes map tag attributes (<fg=..;bg=..;op=..>) to SGR
// parameters; they are used by createCodeFromString for parsing.
var (
	fgColorNames = map[string]string{
		"black": "30", "red": "31", "green": "32", "yellow": "33",
		"blue": "34", "magenta": "35", "cyan": "36", "white": "37",
		"default": "39", "gray": "90", "grey": "90",
	}
	bgColorNames = map[string]string{
		"black": "40", "red": "41", "green": "42", "yellow": "43",
		"blue": "44", "magenta": "45", "cyan": "46", "white": "47",
		"default": "49", "gray": "100", "grey": "100",
	}
	opCodes = map[string]string{
		"bold": "1", "dark": "2", "italic": "3", "underline": "4",
		"blink": "5", "reverse": "7", "concealed": "8",
	}
)

// createCodeFromString parses a tag name or attribute string into an SGR parameter string.
// Supports named styles (<info>) and attribute styles (<fg=red;bg=blue;op=bold>, <fg=167>, <fg=11aa23>).
func (f *Formatter) createCodeFromString(str string) (string, bool) {
	if code, ok := f.styles[str]; ok {
		return code, true
	}
	if !strings.Contains(str, "=") {
		return "", false
	}
	parts := make([]string, 0, 4)
	for _, attr := range strings.Split(str, ";") {
		attr = strings.TrimSpace(attr)
		if attr == "" {
			continue
		}
		kv := strings.SplitN(attr, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := strings.ToLower(kv[0]), kv[1]
		switch key {
		case "fg", "bg":
			if code, ok := parseColorAttr(key, val); ok {
				parts = append(parts, code)
			}
		case "op":
			if code, ok := opCodes[strings.ToLower(val)]; ok {
				parts = append(parts, code)
			}
		}
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, ";"), true
}

// parseColorAttr parses an fg/bg attribute value: named colors, 0-255 256-color,
// and 3/6-digit hexadecimal RGB.
func parseColorAttr(prefix, val string) (string, bool) {
	named := fgColorNames
	if prefix == "bg" {
		named = bgColorNames
	}
	if code, ok := named[strings.ToLower(val)]; ok {
		return code, true
	}
	if n, err := strconv.Atoi(val); err == nil && n >= 0 && n <= 255 {
		if prefix == "bg" {
			return "48;5;" + val, true
		}
		return "38;5;" + val, true
	}
	hex := strings.TrimPrefix(val, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) == 6 {
		r, e1 := strconv.ParseUint(hex[0:2], 16, 8)
		g, e2 := strconv.ParseUint(hex[2:4], 16, 8)
		b, e3 := strconv.ParseUint(hex[4:6], 16, 8)
		if e1 == nil && e2 == nil && e3 == nil {
			if prefix == "bg" {
				return fmt.Sprintf("48;2;%d;%d;%d", r, g, b), true
			}
			return fmt.Sprintf("38;2;%d;%d;%d", r, g, b), true
		}
	}
	return "", false
}

func (f *Formatter) applyCurrentStyle(text string) string {
	if !f.decorated {
		// Non-decorated mode: strip any leftover tags and output plain text.
		return tagStripRegex.ReplaceAllString(text, "")
	}
	if len(text) == 0 {
		return text
	}
	code := f.currentCode()
	if code == "" {
		return tagStripRegex.ReplaceAllString(text, "")
	}
	// fatih/color outputs plain text internally when colors are unsupported (NoColor).
	attrs := sgrAttrs(code)
	if len(attrs) == 0 {
		return tagStripRegex.ReplaceAllString(text, "")
	}
	return color.New(attrs...).Sprint(text)
}

// sgrAttrs parses an SGR parameter string (e.g. "31;44;1") into a list of fatih/color attributes.
func sgrAttrs(code string) []color.Attribute {
	parts := strings.Split(code, ";")
	attrs := make([]color.Attribute, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		attrs = append(attrs, color.Attribute(n))
	}
	return attrs
}

// push pushes a style code onto the stack.
func (f *Formatter) push(code string) { f.stack = append(f.stack, code) }

// pop pops the top style code from the stack.
func (f *Formatter) pop() string {
	if len(f.stack) == 0 {
		return ""
	}
	last := f.stack[len(f.stack)-1]
	f.stack = f.stack[:len(f.stack)-1]
	return last
}

// popTo pops down to the nesting level matching code (handles </name> closing).
func (f *Formatter) popTo(code string) {
	if len(f.stack) == 0 {
		return
	}
	if code == "" {
		f.pop()
		return
	}
	for i := len(f.stack) - 1; i >= 0; i-- {
		if f.stack[i] == code {
			f.stack = f.stack[:i]
			return
		}
	}
	f.pop()
}

// currentCode returns the top style code of the stack.
func (f *Formatter) currentCode() string {
	if len(f.stack) == 0 {
		return ""
	}
	return f.stack[len(f.stack)-1]
}
