package cli

import (
	"strings"
	"testing"
)

func TestNewTableDefaultRender(t *testing.T) {
	tb := NewTable()
	tb.SetHeader([]string{"Name", "Score"}, AlignLeft)
	tb.SetRows([][]string{{"Alice", "90"}, {"Bob", "85"}}, AlignLeft)
	out := tb.Render()
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Alice") || !strings.Contains(out, "90") {
		t.Errorf("render = %q", out)
	}
	if !strings.Contains(out, "+----") {
		t.Errorf("default style should use ascii borders: %q", out)
	}
}

func TestTableBoxStyle(t *testing.T) {
	tb := NewTable()
	tb.SetStyle("box")
	tb.SetHeader([]string{"A"}, AlignLeft)
	tb.SetRows([][]string{{"1"}}, AlignLeft)
	out := tb.Render()
	if !strings.Contains(out, "┌") || !strings.Contains(out, "└") {
		t.Errorf("box style should use unicode: %q", out)
	}
}

func TestTableMarkdownStyle(t *testing.T) {
	tb := NewTable()
	tb.SetStyle("markdown")
	tb.SetHeader([]string{"A", "B"}, AlignLeft)
	tb.SetRows([][]string{{"1", "2"}}, AlignLeft)
	out := tb.Render()
	if !strings.Contains(out, "|---") {
		t.Errorf("markdown style should have separator row: %q", out)
	}
}

func TestTableUnknownStyleFallsBack(t *testing.T) {
	tb := NewTable()
	tb.SetStyle("nonexistent")
	tb.SetHeader([]string{"A"}, AlignLeft)
	tb.SetRows([][]string{{"1"}}, AlignLeft)
	if tb.style != "default" {
		t.Errorf("style = %q, want default", tb.style)
	}
	if out := tb.Render(); out == "" {
		t.Error("should still render")
	}
}

func TestTableAddRowAndAlign(t *testing.T) {
	tb := NewTable()
	tb.SetHeader([]string{"A"}, AlignRight)
	tb.AddRow([]string{"x"}, false)
	tb.AddRow([]string{"y"}, false)
	tb.AddRow([]string{"first"}, true) // insert at the front
	tb.SetCellAlign(AlignCenter)
	out := tb.Render()
	if !strings.Contains(out, "first") || !strings.Contains(out, "x") {
		t.Errorf("render = %q", out)
	}
}

func TestTableAddRowAny(t *testing.T) {
	tb := NewTable()
	tb.SetHeader([]string{"Item", "Qty"}, AlignLeft)
	tb.AddRowAny([]interface{}{"Apple", 3}, false)
	tb.AddRowAny([]interface{}{"Banana", 1.5}, false)
	out := tb.Render()
	if !strings.Contains(out, "3") || !strings.Contains(out, "1.5") {
		t.Errorf("render = %q", out)
	}
}

func TestTableAddSeparatorAndSection(t *testing.T) {
	tb := NewTable()
	tb.SetHeader([]string{"Item"}, AlignLeft)
	tb.AddSection("Fruits")
	tb.AddRow([]string{"Apple"}, false)
	tb.AddSeparator()
	tb.AddSection("Drinks")
	tb.AddRow([]string{"Tea"}, false)
	out := tb.Render()
	if !strings.Contains(out, "Fruits") || !strings.Contains(out, "Drinks") || !strings.Contains(out, "Tea") {
		t.Errorf("render = %q", out)
	}
}

func TestTableEmpty(t *testing.T) {
	tb := NewTable()
	out := tb.Render()
	if out == "" {
		t.Error("empty table should still render borders")
	}
}

func TestTableNoHeader(t *testing.T) {
	tb := NewTable()
	tb.SetRows([][]string{{"a", "b"}}, AlignLeft)
	out := tb.Render()
	if !strings.Contains(out, "a") {
		t.Errorf("render = %q", out)
	}
}

func TestStringWidthAndWideRunes(t *testing.T) {
	if stringWidth("abc") != 3 {
		t.Errorf("ascii width = %d", stringWidth("abc"))
	}
	if stringWidth("中文") != 4 {
		t.Errorf("wide width = %d", stringWidth("中文"))
	}
	if !isWideRune('中') {
		t.Error("a CJK rune should be wide")
	}
	if isWideRune('a') {
		t.Error("a should not be wide")
	}
}

func TestPadAlignments(t *testing.T) {
	if got := pad("x", 3, AlignLeft); got != "x  " {
		t.Errorf("left = %q", got)
	}
	if got := pad("x", 3, AlignRight); got != "  x" {
		t.Errorf("right = %q", got)
	}
	if got := pad("x", 4, AlignCenter); got != " x  " {
		t.Errorf("center = %q", got)
	}
	if got := pad("toolong", 3, AlignLeft); got != "toolong" {
		t.Errorf("no pad = %q", got)
	}
}

func TestFmtSprintAndStringifyRow(t *testing.T) {
	if fmtSprint("keep") != "keep" {
		t.Errorf("fmtSprint string = %q", fmtSprint("keep"))
	}
	if fmtSprint(42) != "42" {
		t.Errorf("fmtSprint int = %q", fmtSprint(42))
	}
	row := stringifyRow([]interface{}{"s", 1, 1.5})
	if len(row) != 3 || row[0] != "s" || row[1] != "1" {
		t.Errorf("stringifyRow = %v", row)
	}
}
