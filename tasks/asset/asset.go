package asset

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
