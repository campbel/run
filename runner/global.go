package runner

import (
	"io"
	"os"
)

type Event struct {
	Message string
}

type GlobalContext struct {
	out io.Writer
	err io.Writer
	in  io.Reader
	bus chan Event
}

func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		out: os.Stdout,
		err: os.Stderr,
		in:  os.Stdin,
		bus: make(chan Event),
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
	return c.out.Write(p)
}

func (c *GlobalContext) Emit(e Event) {
	c.bus <- e
}

func (c *GlobalContext) Events() <-chan Event {
	return c.bus
}

func (c *GlobalContext) Done() {
	close(c.bus)
}
