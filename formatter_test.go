package cli

import (
	"strings"
	"testing"
)

func TestNewFormatterDefaultStyles(t *testing.T) {
	f := NewFormatter()
	for _, name := range []string{"error", "info", "comment", "warning", "success", "question", "highlight"} {
		if !f.HasStyle(name) {
			t.Errorf("missing default style %q", name)
		}
	}
}

func TestFormatterSetAndGetStyle(t *testing.T) {
	f := NewFormatter()
	f.SetStyle("custom", "35")
	if !f.HasStyle("custom") {
		t.Error("custom style not registered")
	}
	code, err := f.GetStyle("custom")
	if err != nil || code != "35" {
		t.Errorf("GetStyle = %q err=%v", code, err)
	}
	if _, err := f.GetStyle("missing"); err == nil {
		t.Error("expected error for missing style")
	}
}

func TestFormatterDecoratedToggle(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(true)
	if !f.IsDecorated() {
		t.Error("should be decorated")
	}
	f.SetDecorated(false)
	if f.IsDecorated() {
		t.Error("should not be decorated")
	}
}

func TestFormatNamedTagDecorated(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(true)
	out := f.Format("<info>hello</info>")
	// Decorated mode should output ANSI codes (contains ESC or at least no raw tags).
	if strings.Contains(out, "<info>") {
		t.Errorf("tag not processed: %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("content missing: %q", out)
	}
}

func TestFormatNamedTagPlain(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(false)
	out := f.Format("<error>boom</error>")
	if out != "boom" {
		t.Errorf("plain format = %q", out)
	}
}

func TestFormatAttrTags(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(true)
	out := f.Format("<fg=red;bg=blue;op=bold>text</>")
	if strings.Contains(out, "fg=red") || !strings.Contains(out, "text") {
		t.Errorf("attr tag output = %q", out)
	}
}

func TestFormatUnknownTagKeepsText(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(false)
	out := f.Format("<nosuchtag>keep</nosuchtag>")
	if !strings.Contains(out, "keep") {
		t.Errorf("unknown tag output = %q", out)
	}
}

func TestFormatEscape(t *testing.T) {
	f := NewFormatter()
	esc := f.Escape("plain <info>not a tag</info>")
	// After escaping, every < should be preceded by a backslash so it is not parsed as a tag.
	if !strings.Contains(esc, `\<info>`) || !strings.Contains(esc, `\</info>`) {
		t.Errorf("escape did not escape: %q", esc)
	}
}

func TestFormatNestedTags(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(false)
	out := f.Format("<info>a <comment>b</comment> c</info>")
	if out != "a b c" {
		t.Errorf("nested format = %q", out)
	}
}

func TestFormatUnbalancedClose(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(false)
	// A surplus closing tag should not panic.
	out := f.Format("text</>")
	if out != "text" {
		t.Errorf("unbalanced = %q", out)
	}
}

func TestCreateCodeFromString(t *testing.T) {
	f := NewFormatter()
	if code, ok := f.createCodeFromString("info"); !ok || code == "" {
		t.Errorf("named code = %q ok=%v", code, ok)
	}
	if _, ok := f.createCodeFromString("noise"); ok {
		t.Error("noise should not resolve")
	}
	if code, ok := f.createCodeFromString("fg=red"); !ok || code == "" {
		t.Errorf("attr code = %q ok=%v", code, ok)
	}
}

func TestFormatRawTextWithPlainContent(t *testing.T) {
	f := NewFormatter()
	f.SetDecorated(false)
	out := f.Format("no tags at all")
	if out != "no tags at all" {
		t.Errorf("plain text = %q", out)
	}
}
