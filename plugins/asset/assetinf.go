package assets

import (
	"os"
	"time"
)

type Asset struct {
	Name    string
	ZBuf    []byte
	Size    int
	Mode    os.FileMode
	modTime time.Time
	Buf     []byte
}
