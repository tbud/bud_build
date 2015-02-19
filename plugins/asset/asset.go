package asset

import (
	. "github.com/tbud/bud/context"
	// "github.com/tbud/x/log"
	// "github.com/tbud/bud/plugins"
)

var _assets = map[string]Asset{}

func Register(assets []Asset) {

	if assets == nil {
		panic("bud: Register assets is nil")
	}

	for _, asset := range assets {
		if _, dup := _assets[asset.Name]; dup {
			panic("bud: Register called twice for asset " + asset.Name)
		}
	}
}

type AssetTask struct {
	Package    string
	BaseDir    string
	Includes   []string
	Excludes   []string
	Output     string
	NoCompress bool
}

func (a *AssetPlugin) Execute() error {
	return nil
}

func (a *AssetPlugin) Validate() error {
	return nil
}

func init() {
	Task("asset", &AssetTask{Package: "main", Includes: []string{"**"}})
}
