package runner

import (
	"io"
	"os"
)

type GlobalContext struct {
	out io.Writer
	err io.Writer
	in  io.Reader
}

func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		out: os.Stdout,
		err: os.Stderr,
		in:  os.Stdin,
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
