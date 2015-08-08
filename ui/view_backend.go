package ui

import (
	"bytes"
	"fmt"

	"github.com/tcolar/goed/actions"
	"github.com/tcolar/goed/core"
)

func (v *View) Save() {
	e := core.Ed
	err := v.backend.Save(v.backend.SrcLoc())
	if err != nil {
		e.SetStatusErr("Saving Failed " + err.Error())
		return
	}
	v.SetDirty(false)
	e.SetStatus("Saved " + v.backend.SrcLoc())
}

// InsertCur inserts text at the current location.
func (v *View) InsertCur(s string) {
	_, y, x := v.CurChar()
	if len(v.selections) > 0 {
		s := v.selections[0]
		v.MoveCursorRoll(s.LineFrom-y, s.ColFrom-x)
		v.SelectionDelete(&s)
		v.ClearSelections()
	}
	_, y, x = v.CurChar()
	v.Insert(y, x, s, true)
}

// Insert inserts text at the given text location
func (v *View) Insert(line, col int, s string, undoable bool) {
	e := core.Ed
	if s == "\n" {
		if col >= v.LineLen(v.slice, line) {
			s += string(v.lineIndent(line))
		}
	}
	err := v.backend.Insert(line, col, s)
	if err != nil {
		e.SetStatusErr("Insert Failed " + err.Error())
		return
	}

	// move the cursor to after insertion
	c := v.lineColsTo(v.slice, line, col)
	b := []byte(s)
	offy := bytes.Count(b, core.LineSep)
	idx := bytes.LastIndex(b, core.LineSep) + 1
	lnLen := len(b[idx:])
	offx := 0
	if offy > 0 {
		offx -= c
	}
	offx += v.strSize(string(b[idx:]))
	if offy == 0 {
		lnLen += col
	}

	if undoable {
		actions.UndoAdd(
			v.Id(),
			actions.NewViewInsertAction(v.Id(), line, col, s, false),
			actions.NewViewDeleteAction(v.Id(), line, col, line+offy, lnLen-1, false))
	}
	v.Render()
	e.TermFlush()
	e.SetStatus(fmt.Sprintf("v %d %d(%d) - %d %d\n", offy, offx, c, line+offy-v.CurLine(), c+offx-v.CurCol()))
	v.MoveCursor(line+offy-v.CurLine(), c+offx-v.CurCol())
}

func (v *View) lineIndent(line int) []rune {
	ln := v.Line(v.slice, line)
	for i, c := range ln {
		if c != ' ' && c != '\t' {
			return ln[:i]
		}
	}
	return ln
}

func (v *View) InsertNewLineCur() {
	v.InsertCur("\n")
}

// InsertNewLine inserts a "newline"(Enter key) in the buffer
func (v *View) InsertNewLine(line, col int) {
	v.Insert(line, col, "\n", true)
}

func (v *View) Reload() {
	err := v.backend.Reload()
	if err != nil {
		core.Ed.SetStatusErr(err.Error())
	}
	actions.UndoClear(v.Id())
	v.Render()
	core.Ed.TermFlush()
}

// Delete removes characters at the given text location
func (v *View) Delete(line1, col1, line2, col2 int, undoable bool) {
	s := core.NewSelection(line1, col1, line2, col2)
	text := core.RunesToString(v.SelectionText(s))
	err := v.backend.Remove(line1, col1, line2, col2)
	if err != nil {
		core.Ed.SetStatusErr("Delete Failed " + err.Error())
		return
	}
	if undoable {
		actions.UndoAdd(
			v.Id(),
			actions.NewViewDeleteAction(v.Id(), line1, col1, line2, col2, false),
			actions.NewViewInsertAction(v.Id(), line1, col1, text, false))
	}
	v.Render()
	core.Ed.TermFlush()
	offx := v.lineColsTo(v.Slice(), line1, col1)
	v.MoveCursor(line1-v.CurLine(), offx-v.CurCol())
}

// DeleteCur removes a selection or the curent character
func (v *View) DeleteCur() {
	c, y, x := v.CurChar()
	if len(v.selections) > 0 {
		s := v.selections[0]
		v.MoveCursorRoll(s.LineFrom-y, s.ColFrom-x)
		v.SelectionDelete(&s)
		v.ClearSelections()
		return
	}
	if c != nil {
		v.Delete(y, x, y, x, true)
	}
}

// Backspace removes a selection or character before the current location
func (v *View) Backspace() {
	if v.CurLine() == 0 && v.CurCol() == 0 {
		return
	}
	if len(v.selections) == 0 {
		v.MoveCursorRoll(0, -1)
	}
	v.DeleteCur()
}

// LineCount return the number of lines in the  buffer
// if the last line is a blank line, do not count it
func (v *View) LineCount() int {
	return v.backend.LineCount()
}

// Line return the line at the given index
func (v *View) Line(s *core.Slice, lnIndex int) []rune {
	// backend is 1-based indexed
	index := lnIndex - s.R1
	if index < 0 || index >= len(*s.Text()) {
		return []rune{}
	}
	return (*s.Text())[index]
}

// LineLen returns the length of a line (raw runes length)
func (v *View) LineLen(s *core.Slice, lnIndex int) int {
	return len(v.Line(s, lnIndex))
}

// LineCol returns the number of columns used for the given lines
// ie: a tab uses multiple columns
func (v *View) lineCols(s *core.Slice, lnIndex int) int {
	return v.lineColsTo(s, lnIndex, v.LineLen(s, lnIndex))
}

// LineColsTo returns the number of columns up to the given line index
// ie: a tab uses multiple columns
func (v *View) lineColsTo(s *core.Slice, lnIndex, to int) int {
	line := v.Line(s, lnIndex)
	if lnIndex > v.LineCount() || to > len(line) {
		return 0
	}
	ln := 0
	for _, r := range line[:to] {
		ln += v.runeSize(r)
	}
	return ln
}

// LineRunesTo returns the number of raw runes to the given line column
func (v View) LineRunesTo(s *core.Slice, lnIndex, column int) int {
	runes := 0
	if len(*s.Text()) == 0 || lnIndex >= v.LineCount() || lnIndex < 0 {
		return 0
	}
	ln := v.Line(s, lnIndex)
	for i := 0; i <= column && runes < len(ln); {
		i += v.runeSize(ln[runes])
		if i <= column {
			runes++
		}
	}
	return runes
}

// CursorChar returns the rune at the given cursor location
// Also returns the position of the char in the text buffer
func (v *View) CursorChar(s *core.Slice, cursorY, cursorX int) (r *rune, textY, textX int) {
	// backend is 1-based indexed
	x, y := v.LineRunesTo(s, cursorY, cursorX), cursorY
	ln := v.Line(s, y)
	if len(ln) <= x { // EOL
		nl := '\n'
		return &nl, y, x
	} else if len(ln) <= x {
		return nil, y, x
	}
	return &ln[x], y, x
}

// CurChar returns the rune at the current cursor location
func (v *View) CurChar() (r *rune, textY, textX int) {
	return v.CursorChar(v.slice, v.CurLine(), v.CurCol())
}

// The runeSize (on screen)
// tabs are a special case
func (v *View) runeSize(r rune) int {
	if r == '\t' {
		return tabSize
	}
	return 1
}

// The string size (on screen)
// tabs are a special case
func (v *View) strSize(s string) int {
	ln := 0
	for _, r := range s {
		ln += v.runeSize(r)
	}
	return ln
}
