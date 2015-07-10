package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tcolar/goed/backend"
	"github.com/tcolar/goed/core"
)

// Cmdbar is the CommandBar widget
// TODO: needs to be decoupled from it's actions
type Cmdbar struct {
	Widget
	Cmd     []rune
	History [][]rune // TODO : cmd history
}

func (c *Cmdbar) Render() {
	ed := core.Ed
	t := ed.Theme()
	ed.TermFB(t.Cmdbar.Fg, t.Cmdbar.Bg)
	ed.TermFill(t.Cmdbar.Rune, c.y1, c.x1, c.y2, c.x2)
	if ed.CmdOn() {
		ed.TermFB(t.CmdbarTextOn, t.Cmdbar.Bg)
		ed.TermStr(c.y1, c.x1, fmt.Sprintf("> %s", string(c.Cmd)))
	} else {
		ed.TermFB(t.CmdbarText, t.Cmdbar.Bg)
		ed.TermStr(c.y1, c.x1, fmt.Sprintf("> %s", string(c.Cmd)))
	}
	ed.TermFB(t.CmdbarText, t.Cmdbar.Bg)
	ed.TermStr(c.y1, c.x2-11, fmt.Sprintf("|GoEd %s", core.Version))
}

func (c *Cmdbar) RunCmd() {
	ed := core.Ed.(*Editor)
	// TODO: This is temporary until I create real fs based events & actions
	s := string(c.Cmd)
	parts := strings.Fields(s)
	if len(parts) < 1 {
		return
	}
	args := parts[1:]
	var err error
	switch parts[0] {
	//case "d", "del": // as vi del
	//case "dd":
	case "dc", "delcol":
		ed.DelColCheck(ed.CurCol)
	case "dv", "delview":
		ed.DelViewCheck(ed.curView)
	case "e", "exec":
		c.exec(args)
	//case "h", "help":
	//	ed.SetStatus("TBD help")
	case "nc", "newcol":
		c.newCol(args)
	case "nv", "newview":
		c.newView(args)
	case "o", "open":
		err = c.open(args)
	case "p", "paste": // as vi
		c.paste(args)
	case "s", "save":
		c.save(args)
	case ":", "line":
		c.line(args)
	case "/", "search":
		if len(c.Cmd) < 2 {
			break
		}
		query := string(c.Cmd[2:])
		c.Search(query)
	case "y", "yank": // as vi copy
		err = c.yank(args)
	case "yy":
		err = c.yank([]string{"1"})
	default:
		ed.SetStatusErr("Unexpected command " + parts[0])
	}
	if err == nil {
		ed.cmdOn = false
	} else {
		ed.SetStatus(err.Error())
	}
}

func (c *Cmdbar) paste(args []string) {
	ed := core.Ed.(*Editor)
	v := ed.curView
	v.MoveCursorRoll(1, -v.CurCol())
	l := v.CurLine()
	v.Paste()
	x, y := v.CurCol(), v.CurLine()
	v.Insert(y, x, "\n")
	v.MoveCursorRoll(l-y, -x)
	v.SetDirty(true)
}

func (c *Cmdbar) yank(args []string) error {
	ed := core.Ed.(*Editor)
	v := ed.curView
	if len(args) == 0 {
		return fmt.Errorf("Expected an argument")
	}
	nb, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("Expected a numericargument")
	}
	nb--
	s := &core.Selection{
		LineFrom: v.CurLine(),
		LineTo:   v.CurLine() + nb,
		ColTo:    -1,
	}
	ed.curView.SelectionCopy(s)
	return nil
}

func (c *Cmdbar) open(args []string) error {
	if len(args) < 1 {
		// try to expand a location from the current view
		return fmt.Errorf("No path provided")
	}
	ed := core.Ed.(*Editor)
	v := ed.NewView(args[0])
	_, err := ed.Open(args[0], v, ed.CurView().WorkDir())
	if err != nil {
		return err
	}
	ed.InsertViewSmart(v)
	ed.ActivateView(v, 0, 0)
	return nil
}

func (c *Cmdbar) save(args []string) {
	ed := core.Ed.(*Editor)
	if ed.CurView != nil {
		ed.curView.Save()
	}
}

func (c *Cmdbar) line(args []string) {
	ed := core.Ed.(*Editor)
	if len(args) < 0 {
		ed.SetStatusErr("Expected a line number argument.")
		return
	}
	l, err := strconv.Atoi(args[0])
	if err != nil {
		ed.SetStatusErr("Expected a line number argument.")
		return
	}
	if ed.CurView != nil {
		ed.curView.MoveCursor(l-ed.curView.CurLine()-1, 0)
	}
}

func (c *Cmdbar) Search(query string) {
	c.exec([]string{"grep", "-rn", query})
}

func (c *Cmdbar) newCol(args []string) {
	// nc : newblank col
	// nc [file], new col, open file
	// nc 40 -> new col 40% width
	// nc 40 [file] -> new col 40% width, open file
	loc := ""
	pct := 50
	if len(args) > 0 {
		p, err := strconv.Atoi(args[0])
		if err == nil {
			pct = p
			if len(args) > 1 {
				loc = strings.Join(args[1:], " ")
			}
		} else {
			loc = strings.Join(args, " ")
		}
	}
	ed := core.Ed.(*Editor)
	v := ed.AddCol(ed.CurCol, float64(pct)/100.0).Views[0]
	if len(loc) > 0 {
		ed.Open(loc, v, "")
	}
}

func (c *Cmdbar) newView(args []string) {
	loc := ""
	pct := 50
	if len(args) > 0 {
		p, err := strconv.Atoi(args[0])
		if err == nil {
			pct = p
			if len(args) > 1 {
				loc = strings.Join(args[1:], " ")
			}
		} else {
			loc = strings.Join(args, " ")
		}
	}
	ed := core.Ed.(*Editor)
	v := ed.AddView(ed.curView, float64(pct)/100.0)
	if len(loc) > 0 {
		ed.Open(loc, v, "")
	}
}

func (c *Cmdbar) exec(args []string) {
	ed := core.Ed.(*Editor)
	workDir := "."
	if ed.curView != nil {
		workDir = ed.CurView().WorkDir()
	}
	v := ed.AddViewSmart()
	b, err := backend.NewMemBackendCmd(args, workDir, v.Id(), nil)
	if err != nil {
		ed.SetStatusErr(err.Error())
	}
	v.backend = b
}
