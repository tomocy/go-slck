package tcp

import (
	"flag"
	"fmt"
	"io"
)

type app struct {
	w    io.Writer
	addr string
}

func (a *app) parseFlags(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("too less arguments")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	addr := flags.String("addr", ":80", "the address to listen and serve")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	a.addr = *addr

	return nil
}

func (a *app) printf(format string, as ...interface{}) {
	fmt.Fprintf(a.w, format, as...)
}
