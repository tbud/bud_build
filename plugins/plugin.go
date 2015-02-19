package plugins

type Plugin interface {
	Execute() error
	Validate() error
}
