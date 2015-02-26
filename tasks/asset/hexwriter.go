package asset

import (
	"io"
)

const hextable = "0123456789abcdef"

type HexWriter struct {
	io.Writer
	off int
}

func (w *HexWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	buf := []byte(`\x00`)
	var b byte

	for n, b = range p {
		buf[2] = hextable[b>>4]
		buf[3] = hextable[b&0x0f]
		w.Writer.Write(buf)
		w.off++
	}

	n++

	return
}
