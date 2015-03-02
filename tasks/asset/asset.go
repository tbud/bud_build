package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	. "github.com/tbud/bud/context"
	"io"
	"os"
	"time"
)

var _assets = map[string]Asset{}

func Register(assets []Asset) {
	if assets == nil {
		panic("bud: Register assets is nil")
	}

	for _, asset := range assets {
		if _, dup := _assets[asset.N]; dup {
			panic("bud: Register called twice for asset " + asset.N)
		} else {
			_assets[asset.N] = asset
		}
	}
}

func Open(name string) (rc io.ReadCloser, err error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("asset open file name empty")
	}

	if asset, exist := _assets[name]; exist {
		Log.Debug("Open '%s' from asset.", name)
		return &asset, nil
	} else {
		Log.Debug("Open '%s' from file system.", name)
		return os.Open(name)
	}
}

func Stat(name string) (fi os.FileInfo, err error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("asset open file name empty")
	}

	if asset, exist := _assets[name]; exist {
		Log.Debug("Load '%s' state from asset.", name)
		return &asset, nil
	} else {
		Log.Debug("Load '%s' state from file system.", name)
		return os.Stat(name)
	}
}

type Asset struct {
	N   string      // save name
	Z   []byte      // save compressed data
	S   int         // origin file size
	M   os.FileMode // file mode
	MT  time.Time   // file mod time
	Buf *bytes.Buffer
}

func (a *Asset) Read(p []byte) (n int, err error) {
	if a.Buf == nil {
		gz, err := gzip.NewReader(bytes.NewBuffer(a.Z))
		if err != nil {
			return 0, err
		}

		a.Buf = bytes.NewBuffer(nil)
		_, err = io.Copy(a.Buf, gz)
		gz.Close()
		if err != nil {
			return 0, err
		}

		if a.Buf.Len() != a.S {
			return 0, fmt.Errorf("Asset unzip file %s error.", a.N)
		}
	}

	return a.Buf.Read(p)
}

func (a *Asset) Close() error {
	return nil
}

func (a *Asset) Name() string {
	return a.N
}

func (a *Asset) Size() int64 {
	return int64(a.S)
}

func (a *Asset) Mode() os.FileMode {
	return a.M
}

func (a *Asset) ModTime() time.Time {
	return a.MT
}

func (a *Asset) IsDir() bool {
	return false
}
func (a *Asset) Sys() interface{} {
	return nil
}

const assetTemplate = `
package {{ .assetTask.Package }}

import (
	"time"
	"github.com/tbud/bud/tasks/asset"
)	

func init() {
	asset.Register([]asset.Asset{
		{{ range $asset := .assets }}{
			N: "{{ $asset.N }}",
			Z: []byte("{{ zipFile $asset.N $.assetTask.BaseDir | printf "%s" }}"),
			S: {{ $asset.S }},
			M: {{ printf "%d" $asset.M }},
			MT : time.Unix({{ $asset.MT.Unix }}, 0),
		},{{ end }}
	})
}
`
