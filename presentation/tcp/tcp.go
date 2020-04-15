package tcp

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/tomocy/slck"
)

func New(w io.Writer) *app {
	return &app{
		w: w,
	}
}

type app struct {
	w    io.Writer
	addr string
}

func (a app) Run(args []string) error {
	if err := a.parseFlags(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	ctx := context.TODO()
	if err := a.listenAndServe(ctx); err != nil {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}

	return nil
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
	a.printf("listen and serve on %s\n", a.addr)
	lis, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		registered = make(chan slck.Client)
		deleted    = make(chan slck.Client)
	)
	defer func() {
		close(registered)
		close(deleted)
	}()

	w := slck.NewWorkplace(registered)
	go w.Listen(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := lis.Accept()
			if err != nil {
				return fmt.Errorf("failed to accept connection: %w", err)
			}

			cli := slck.NewClient(conn, registered, deleted)
			go cli.Listen(ctx)
		}
	}
}

func (a *app) printf(format string, as ...interface{}) {
	fmt.Fprintf(a.w, format, as...)
}
