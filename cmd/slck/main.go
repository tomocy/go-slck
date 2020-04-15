package main

import (
	"os"

	"github.com/tomocy/slck/presentation/tcp"
)

func main() {}

func newRunner() runner {
	return tcp.New(os.Stdout)
}

type runner interface {
	Run([]string) error
}
