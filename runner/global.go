package runner

import (
	"io"
	"os"

	"github.com/campbel/run/types"
)

type GlobalContext struct {
	out io.Writer
	err io.Writer
	in  io.Reader
	bus chan types.EventMsg
}

func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		out: os.Stdout,
		err: os.Stderr,
		in:  os.Stdin,
		bus: make(chan types.EventMsg),
	}
}

func (c *GlobalContext) WithStdout(out io.Writer) *GlobalContext {
	c.out = out
	return c
}

func (c *GlobalContext) WithErrout(err io.Writer) *GlobalContext {
	c.err = err
	return c
}

func (c *GlobalContext) WithStdin(in io.Reader) *GlobalContext {
	c.in = in
	return c
}

func (c *GlobalContext) Write(p []byte) (n int, err error) {
	c.bus <- types.EventMsg{
		EventType: types.EventTypeOutput,
		Message:   string(p),
	}
	return len(p), nil
}

func (c *GlobalContext) Emit(e types.EventMsg) {
	c.bus <- e
}

func (c *GlobalContext) Events() <-chan types.EventMsg {
	return c.bus
}

func (c *GlobalContext) Done() {
	close(c.bus)
}
