package tcp

import "io"

type app struct {
	w    io.Writer
	addr string
}
