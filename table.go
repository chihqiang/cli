package cli

import (
	"fmt"
	"strings"
)

// Table alignment modes.
const (
	AlignLeft   = 1
	AlignRight  = 0
	AlignCenter = 2
)

// tableRowKind distinguishes data rows, separators, and section titles.
type tableRowKind int

const (
	rowData tableRowKind = iota
	rowSeparator
	rowSection
)

// tableRow is the render data of a single row.
type tableRow struct {
	cells []string
	kind  tableRowKind
	title string
}

// Table renders a table to the console.
type Table struct {
	header      []string
	headerAlign int
	rows        []tableRow
	cellAlign   int
	colWidth    map[int]int
	style       string
	format      map[string][6][4]string
}

// NewTable creates a table.
func NewTable() *Table {
	t := &Table{
		headerAlign: AlignLeft,
		cellAlign:   AlignLeft,
		style:       "default",
		colWidth:    make(map[int]int),
	}
	t.format = map[string][6][4]string{
		"compact": {},
		"default": {
			{"+", "-", "+", "+"},
			{"|", " ", "|", "|"},
			{"+", "-", "+", "+"},
			{"+", "-", "+", "+"},
			{"+", "-", "-", "+"},
			{"+", "-", "-", "+"},
		},
		"markdown": {
			{" ", " ", " ", " "},
			{"|", " ", "|", "|"},
			{"|", "-", "|", "|"},
			{" ", " ", " ", " "},
			{"|", " ", " ", "|"},
			{"|", " ", " ", "|"},
		},
		"borderless": {
			{"=", "=", " ", "="},
			{" ", " ", " ", " "},
			{"=", "=", " ", "="},
			{"=", "=", " ", "="},
			{"=", "=", " ", "="},
			{"=", "=", " ", "="},
		},
		"box": {
			{"┌", "─", "┬", "┐"},
			{"│", " ", "│", "│"},
			{"├", "─", "┼", "┤"},
			{"└", "─", "┴", "┘"},
			{"├", "─", "┴", "┤"},
			{"├", "─", "┬", "┤"},
		},
		"box-double": {
			{"╔", "═", "╤", "╗"},
			{"║", " ", "│", "║"},
			{"╠", "─", "╪", "╣"},
			{"╚", "═", "╧", "╝"},
			{"╠", "═", "╧", "╣"},
			{"╠", "═", "╤", "╣"},
		},
	}
	return t
}

// SetHeader sets the table header.
func (t *Table) SetHeader(header []string, align int) *Table {
	t.header = header
	t.headerAlign = align
	t.checkColWidth(header)
	return t
}

// SetRows replaces the data rows.
func (t *Table) SetRows(rows [][]string, align int) *Table {
	t.rows = nil
	for _, r := range rows {
		t.AddRow(r, false)
	}
	t.cellAlign = align
	return t
}

// SetCellAlign sets the cell alignment.
func (t *Table) SetCellAlign(align int) *Table {
	t.cellAlign = align
	return t
}

// AddRow appends a row.
func (t *Table) AddRow(row []string, first bool) *Table {
	tr := tableRow{kind: rowData, cells: append([]string{}, row...)}
	if first {
		t.rows = append([]tableRow{tr}, t.rows...)
	} else {
		t.rows = append(t.rows, tr)
	}
	t.checkColWidth(row)
	return t
}

// AddRowAny appends a row with heterogeneous cell types.
func (t *Table) AddRowAny(row []interface{}, first bool) *Table {
	return t.AddRow(stringifyRow(row), first)
}

// AddSeparator appends a horizontal separator line.
func (t *Table) AddSeparator() *Table {
	t.rows = append(t.rows, tableRow{kind: rowSeparator})
	return t
}

// AddSection appends a full-width section title row.
func (t *Table) AddSection(title string) *Table {
	t.rows = append(t.rows, tableRow{kind: rowSection, title: title})
	return t
}

// SetStyle sets the render style.
func (t *Table) SetStyle(style string) *Table {
	if _, ok := t.format[style]; ok {
		t.style = style
	}
	return t
}

func (t *Table) checkColWidth(row []string) {
	for k, cell := range row {
		w := stringWidth(strings.TrimSpace(cell))
		if _, ok := t.colWidth[k]; !ok || w > t.colWidth[k] {
			t.colWidth[k] = w
		}
	}
}

func (t *Table) getStyle(pos string) [4]string {
	if f, ok := t.format[t.style]; ok {
		return f[posIndex(pos)]
	}
	return [4]string{" ", " ", " ", " "}
}

func posIndex(pos string) int {
	switch pos {
	case "top":
		return 0
	case "cell":
		return 1
	case "middle":
		return 2
	case "bottom":
		return 3
	case "cross-top":
		return 4
	case "cross-bottom":
		return 5
	}
	return 1
}

func (t *Table) renderSeparator(pos string) string {
	style := t.getStyle(pos)
	parts := make([]string, 0)
	for i := 0; i < len(t.colWidth); i++ {
		parts = append(parts, strings.Repeat(style[1], t.colWidth[i]+2))
	}
	if len(parts) == 0 {
		return style[0] + style[3] + "\n"
	}
	return style[0] + strings.Join(parts, style[2]) + style[3] + "\n"
}

func (t *Table) renderHeader() string {
	style := t.getStyle("cell")
	content := t.renderSeparator("top")

	if len(t.header) == 0 {
		return content
	}

	array := make([]string, 0)
	for k, h := range t.header {
		w := t.colWidth[k]
		array = append(array, " "+pad(h, w, t.headerAlign))
	}
	content += style[0] + strings.Join(array, " "+style[2]) + " " + style[3] + "\n"
	if len(t.rows) > 0 {
		content += t.renderSeparator("middle")
	}
	return content
}

// Render returns the table text.
func (t *Table) Render() string {
	content := t.renderHeader()
	style := t.getStyle("cell")

	for _, row := range t.rows {
		switch row.kind {
		case rowSeparator:
			content += t.renderSeparator("middle")
		case rowSection:
			content += t.renderSeparator("cross-top")
			width := 3*(len(t.colWidth)-1) + sumWidths(t.colWidth)
			content += style[0] + " " + pad(row.title, width, AlignLeft) + " " + style[3] + "\n"
			content += t.renderSeparator("cross-bottom")
		default:
			array := make([]string, 0)
			for k := 0; k < len(t.colWidth); k++ {
				val := ""
				if k < len(row.cells) {
					val = row.cells[k]
				}
				w := t.colWidth[k]
				array = append(array, " "+pad(val, w, t.cellAlign))
			}
			if len(array) == 0 {
				content += style[0] + style[3] + "\n"
			} else {
				content += style[0] + strings.Join(array, " "+style[2]) + " " + style[3] + "\n"
			}
		}
	}

	content += t.renderSeparator("bottom")
	return content
}

func sumWidths(m map[int]int) int {
	s := 0
	for _, v := range m {
		s += v
	}
	return s
}

// ------------------------- Rendering helpers -------------------------

// fmtSprint keeps strings as-is and uses fmt.Sprint for everything else.
func fmtSprint(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// stringifyRow converts a heterogeneous cell row into a string slice.
func stringifyRow(row []interface{}) []string {
	out := make([]string, len(row))
	for i, v := range row {
		out[i] = fmtSprint(v)
	}
	return out
}

// stringWidth computes the display width (East Asian wide characters count as 2).
func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		if r == 0 {
			continue
		}
		if isWideRune(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isWideRune(r rune) bool {
	return r >= 0x1100 && (r <= 0x115F ||
		r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0xA4CF) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE30 && r <= 0xFE4F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x1F300))
}

// pad pads text to the given width using the alignment mode.
func pad(text string, width, align int) string {
	w := stringWidth(text)
	if w >= width {
		return text
	}
	padWidth := width - w
	switch align {
	case AlignRight:
		return strings.Repeat(" ", padWidth) + text
	case AlignCenter:
		left := padWidth / 2
		right := padWidth - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", padWidth)
	}
}
