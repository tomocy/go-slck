package tcp

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/tomocy/slck"
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

func (a app) listenAndServe(ctx context.Context) error {
	lis, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := lis.Accept()
			if err != nil {
				return fmt.Errorf("failed to accept connection: %w", err)
			}

			cli := slck.NewClient(conn)
			go cli.Listen(ctx)
		}
	}
}

func (a *app) printf(format string, as ...interface{}) {
	fmt.Fprintf(a.w, format, as...)
}
