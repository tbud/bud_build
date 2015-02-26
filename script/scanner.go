package script

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const (
	// Continue.
	scanContinue  = iota // uninteresting byte
	scanSkipSpace        // space byte; can skip; known to be last "continue" result
	scanAppendBuf        // byte need to append buf

	// Stop.
	scanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	scanError // hit an error, scanner.err.
)

const (
	bufInImport = iota
	bufInType
	bufInFun
	bufInLine
)

type keywordScanner struct {
	maxWordLen   int
	keywords     []string
	keywordsType func(*scriptScanner, int) int
}

var keywords = keywordScanner{
	0,
	[]string{"import", "func", "type"},
	[]func(*scriptScanner, int) int{stateImport, stateFunc, stateType},
}

func (k *keywordScanner) init() {
	for _, keyword := range k.keywords {
		wordLen := len(keyword)
		if wordLen > k.maxWordLen {
			k.maxWordLen = wordLen
		}
	}
}

func (k *keywordScanner) checkKeyword(s *fileScanner, buf []byte) {
	if len(buf) <= k.maxWordLen {
		value := string(buf)
		for i, keyword := range k.keywords {
			if value == keyword {
				s.bufType = k.keywordsType[i]
				return
			}
		}
	}
}

func init() {
	keywords.init()
}

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

	data         []byte // store data load from file
	parseBuf     []byte // save parsed key or value
	bufType      int    // buf type
	currentState int    // save current state

	Imports []string
	Funcs   []string
	Types   []string
	Lines   []string
}

func (s *scriptScanner) init() {
	s.step = stateBegin
	s.err = nil
	s.bytes = 0
	s.currentState = parseKey
	s.bufInQuote = false
}

func (s *scriptScanner) checkValid(fileName string) (err error) {
	s.init()

	if !filepath.IsAbs(fileName) {
		return fmt.Errorf("file '%s' is not absolute path", fileName)
	}

	s.data, err = ioutil.ReadAll(fileName)
	if err != nil {
		return err
	}

	for _, c := range s.data {
		s.bytes++
		switch s.step(s, int(c)) {
		case scanError:
			return s.err
		case scanAppendBuf:
			s.parseBuf = append(s.parseBuf, c)
		}
	}

	if s.step(s, '\n') == scanError {
		return s.err
	}

	return nil
}

func stateBegin(s *scriptScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}

	switch c {
	case '#':
		s.step = stateComment
		return scanContinue
	}

	s.step = stateLine
	return scanContinue
}

func stateLine(s *scriptScanner, c int) int {
	switch c {
	case '#':
		s.step = stateComment
		return scanContinue
	}
}

func stateError(s *scriptScanner, c int) int {
	return scanError
}

func (s *fileScanner) error(c int, context string) int {
	s.step = stateError
	s.err = fmt.Errorf("invalid character '%c' , Error : %s", c, context)
	return scanError
}

func stateComment(s *scriptScanner, c int) int {
	if c == '\n' || c == '\r' {
		s.step = stateBegin
		return scanContinue
	}
	return scanContinue
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
