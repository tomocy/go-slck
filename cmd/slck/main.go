package main

import (
	"fmt"
	"os"

	"github.com/tomocy/slck/presentation/tcp"
)

func main() {
	if err := newRunner().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRunner() runner {
	return tcp.New(os.Stdout)
}

type runner interface {
	Run([]string) error
}
