package script

type scriptScanner struct {
	// The step is a func to be called to execute the next transition.
	// Also tried using an integer constant and a single func
	// with a switch, but using the func directly was 10% faster
	// on a 64-bit Mac Mini, and it's nicer to read.
	step func(*scriptScanner, int) int

	// Error that happened, if any.
	err error

	// total bytes consumed, updated by decoder.Decode
	bytes int64

	data         []byte   // store data load from file
	baseKeys     []string // save base key
	keyStack     []int    // stack for key
	parseBuf     []byte   // save parsed key or value
	bufType      int      // buf type
	bufInQuote   bool     // when buf type is no quote string and parse in quote then true
	currentState int      // save current state
	kvs          []kvPair //save key value
}
