package backend

import (
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/tcolar/goed/core"
)

func BufferFile(id int) string {
	return path.Join(core.Home, "buffers", strconv.Itoa(id))
}

// Cmd runner with File based backend
// if title == nil then will show the command name
func NewFileBackendCmd(args []string, dir string, viewId int, title *string) (*BackendCmd, error) {
	loc := BufferFile(viewId)
	os.Remove(loc)
	b, err := NewFileBackend(loc, viewId)
	if err != nil {
		return nil, err
	}
	c, err := newBackendCmd(args, dir, viewId, title)
	if err != nil {
		return nil, err
	}
	c.Backend = b
	c.Starter = &FileCmdStarter{}

	go c.Start()
	return c, nil
}

// Cmd runner with In-memory based backend
// if title == nil then will show the command name
func NewMemBackendCmd(args []string, dir string, viewId int, title *string) (*BackendCmd, error) {
	b, err := NewMemBackend("", viewId)
	if err != nil {
		return nil, err
	}
	c, err := newBackendCmd(args, dir, viewId, title)
	if err != nil {
		return nil, err
	}
	c.Backend = b
	c.Starter = &MemCmdStarter{}
	go c.Start()
	return c, nil
}

func newBackendCmd(args []string, dir string, viewId int, title *string) (*BackendCmd, error) {
	c := &BackendCmd{
		dir:    dir,
		runner: exec.Command(args[0], args[1:]...),
		title:  title,
	}
	c.runner.Dir = dir
	if c.title == nil {
		title := strings.Join(c.runner.Args, " ")
		c.title = &title
	}
	return c, nil
}
