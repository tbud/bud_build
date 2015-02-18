package assets

import (
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/log"
)

var _assets = map[string]Asset{}

func Register(assets []Asset) {

	if assets == nil {
		panic("bud: Register assets is nil")
	}

	for _, asset := range assets {
		if _, dup := assets[name]; dup {
			panic("bud: Register called twice for asset " + asset.Name)
		}
	}
}
