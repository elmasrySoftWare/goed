package ui

import (
	"bytes"
	"unicode/utf8"

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
	selections := v.Selections()
	cl, cc := v.CurTextPos()
	v.SetDirty(true)
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
	b := []byte(s)
	endLn := line + bytes.Count(b, core.LineSep)
	idx := bytes.LastIndex(b, core.LineSep) + 1
	endCol := utf8.RuneCount(b[idx:])
	if line == endLn {
		endCol += col
	}

	if undoable {
		actions.UndoAdd(
			v.Id(),
			[]core.Action{
				actions.NewViewInsertAction(v.Id(), line, col, s, false),
				actions.NewSetCursorAction(v.Id(), endLn, endCol)},
			append([]core.Action{
				actions.NewViewDeleteAction(v.Id(), line, col, endLn, endCol-1, false),
				actions.NewSetCursorAction(v.Id(), cl, cc)},
				actions.NewSetSelectionsActions(v.Id(), selections)...),
		)
	}
	v.Render()
	e.TermFlush()
	v.SetCursorPos(endLn, endCol)
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
	cl, cc := v.CurTextPos()
	selections := v.Selections()
	v.SetDirty(true)
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
			[]core.Action{
				actions.NewViewDeleteAction(v.Id(), line1, col1, line2, col2, false),
				actions.NewSetCursorAction(v.Id(), line1, col1)},
			append([]core.Action{
				actions.NewViewInsertAction(v.Id(), line1, col1, text, false),
				actions.NewSetCursorAction(v.Id(), cl, cc)},
				actions.NewSetSelectionsActions(v.Id(), selections)...),
		)
	}
	v.Render()
	core.Ed.TermFlush()
	v.SetCursorPos(line1, col1)
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
func (v *View) Line(slice *core.Slice, lnIndex int) []rune {
	s := slice
	if !s.ContainsLine(lnIndex) {
		s = v.backend.Slice(lnIndex, 0, lnIndex, -1)
	}
	index := lnIndex - s.R1
	if index < 0 || index >= len(*s.Text()) {
		return []rune{}
	}
	return (*s.Text())[index]
}

// LineLen returns the length onf a line (raw runes length)
func (v *View) LineLen(slice *core.Slice, lnIndex int) int {
	s := slice
	if !s.ContainsLine(lnIndex) {
		s = v.backend.Slice(lnIndex, 0, lnIndex, -1)
	}
	return len(v.Line(s, lnIndex))
}

// LineCol returns the number of columns used for the given lines
// ie: a tab uses multiple columns
func (v *View) lineCols(slice *core.Slice, lnIndex int) int {
	s := slice
	if !s.ContainsLine(lnIndex) {
		s = v.backend.Slice(lnIndex, 0, lnIndex, -1)
	}
	return v.lineColsTo(s, lnIndex, v.LineLen(s, lnIndex))
}

// LineColsTo returns the number of columns up to the given line index
// ie: a tab uses multiple columns
func (v *View) lineColsTo(s *core.Slice, lnIndex, to int) int {
	if lnIndex > v.LineCount() {
		return 0
	}
	line := v.Line(s, lnIndex)
	if len(line) == 0 {
		return 0
	}
	ln := 0
	for i := 0; i < to && i < len(line); i++ {
		ln += v.runeSize(line[i])
	}
	return ln
}

// LineRunesTo returns the number of raw runes to the given line column
func (v View) LineRunesTo(slice *core.Slice, lnIndex, column int) int {
	s := slice
	if !s.ContainsLine(lnIndex) {
		s = v.backend.Slice(lnIndex, 0, lnIndex, -1)
	}
	runes := 0
	if lnIndex < 0 || lnIndex > v.LineCount() {
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
// Also returns the position of the char in the text buffer (text position)
func (v *View) CursorChar(slice *core.Slice, cursorY, cursorX int) (r *rune, textY, textX int) {
	s := slice
	if !s.ContainsLine(cursorY) {
		s = v.backend.Slice(cursorY, 0, cursorY, -1)
	}
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

// The runeSize (**on screen**)
// tabs are a special case as well as some Asian pictograms
func (v *View) runeSize(r rune) int {
	// variable tab width
	if r == '\t' {
		return tabSize
	}
	// various Asian chars that are printed "double wide" (2 term cells)
	if r >= 0x1100 &&
		(r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff00 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	// "normal" chars
	return 1
}

// The string size (**on screen**)
// tabs are a special case
func (v *View) strSize(s string) int {
	ln := 0
	for _, r := range s {
		ln += v.runeSize(r)
	}
	return ln
}
