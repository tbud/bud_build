package script

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type keywordScanner struct {
	maxWordLen   int
	keywords     []string
	keywordsType []int
}

const (
	bufInUnknown = iota
	bufInImport
	bufInType
	bufInFun
	bufInConst
	bufInVar
	bufInLine
)

var keywords = keywordScanner{
	0,
	[]string{"import", "func", "type", "const", "var"},
	[]int{bufInImport, bufInFun, bufInType, bufInConst, bufInVar},
}

func (k *keywordScanner) init() {
	for _, keyword := range k.keywords {
		wordLen := len(keyword)
		if wordLen > k.maxWordLen {
			k.maxWordLen = wordLen
		}
	}
}

func (k *keywordScanner) checkKeyword(s *scriptScanner, buf []byte) {
	if len(buf) <= k.maxWordLen {
		value := string(buf)
		for i, keyword := range k.keywords {
			if value == keyword {
				s.bufType = k.keywordsType[i]
				return
			}
		}
	}

	s.bufType = bufInLine
	return
}

func init() {
	keywords.init()
}

const (
	// Continue.
	scanContinue  = iota // uninteresting byte
	scanSkipSpace        // space byte; can skip; known to be last "continue" result
	scanAppendBuf        // byte need to append buf

	// Stop.
	scanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	scanError // hit an error, scanner.err.
)

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

	data       []byte // store data load from file
	parseBuf   []byte // save parsed key or value
	bufType    int    // buf type
	bufInQuote bool   // if buf in Quote
	bracketNum int    // save bracket num

	Imports []string
	Funcs   []string
	Types   []string
	Consts  []string
	Vars    []string
	Lines   []string
}

func (s *scriptScanner) init() {
	s.step = stateBegin
	s.err = nil
	s.bytes = 0
	s.bufType = bufInUnknown
	s.bufInQuote = false
	s.bracketNum = 0
}

func (s *scriptScanner) checkValid(fileName string) (err error) {
	s.init()

	if !filepath.IsAbs(fileName) {
		return fmt.Errorf("file '%s' is not absolute path", fileName)
	}

	s.data, err = ioutil.ReadFile(fileName)
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

	s.step = stateParseLine
	return stateParseLine(s, c)
}

func stateParseLine(s *scriptScanner, c int) int {
	if s.bufType == bufInUnknown {
		switch c {
		case ' ':
			keywords.checkKeyword(s, s.parseBuf)
			return scanAppendBuf
		case '"':
			s.bufInQuote = true
			keywords.checkKeyword(s, s.parseBuf)
			return scanAppendBuf
		case '(', '{':
			s.bracketNum += 1
			keywords.checkKeyword(s, s.parseBuf)
			return scanAppendBuf
		case ':', '=':
			s.bufType = bufInLine
			return scanAppendBuf
		}
	} else {
		if s.bufInQuote {
			switch c {
			default:
				return scanAppendBuf
			case '"':
				s.bufInQuote = false
				return scanAppendBuf
			case '\\':
				s.step = stateInStringEsc
				return scanAppendBuf
			}
		} else {
			switch c {
			default:
				return scanAppendBuf
			case '"':
				s.bufInQuote = true
				return scanAppendBuf
			case '(', '{':
				s.bracketNum += 1
				return scanAppendBuf
			case ')', '}', '\r', '\n':
				if c == ')' || c == '}' {
					s.bracketNum -= 1
				}

				if s.bracketNum == 0 {
					s.step = stateEnd
					return scanAppendBuf
				}
				return scanAppendBuf
			}
		}
	}

	return scanAppendBuf
}

func stateEnd(s *scriptScanner, c int) int {
	switch s.bufType {
	case bufInUnknown:
		return s.error(c, "unkown script line: "+string(s.parseBuf))
	case bufInImport:
		s.Imports = append(s.Imports, strings.TrimSpace(string(s.parseBuf)))
	case bufInFun:
		if bytes.Contains(s.parseBuf, []byte("{")) {
			s.Funcs = append(s.Funcs, strings.TrimSpace(string(s.parseBuf)))
		} else {
			s.step = stateParseLine
			return scanAppendBuf
		}
	case bufInType:
		s.Types = append(s.Types, strings.TrimSpace(string(s.parseBuf)))
	case bufInConst:
		s.Consts = append(s.Consts, strings.TrimSpace(string(s.parseBuf)))
	case bufInVar:
		s.Vars = append(s.Vars, strings.TrimSpace(string(s.parseBuf)))
	case bufInLine:
		s.Lines = append(s.Lines, strings.TrimSpace(string(s.parseBuf)))
	}

	s.parseBuf = s.parseBuf[:0]
	s.bufType = bufInUnknown
	s.step = stateBegin
	return stateBegin(s, c)
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func stateInStringEsc(s *scriptScanner, c int) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		s.step = stateParseLine
		return scanAppendBuf
	}
	if c == 'u' {
		s.step = stateInStringEscU
		return scanAppendBuf
	}
	return s.error(c, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` during a quoted string.
func stateInStringEscU(s *scriptScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU1
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU1 is the state after reading `"\u1` during a quoted string.
func stateInStringEscU1(s *scriptScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU12
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU12 is the state after reading `"\u12` during a quoted string.
func stateInStringEscU12(s *scriptScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU123
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU123 is the state after reading `"\u123` during a quoted string.
func stateInStringEscU123(s *scriptScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateParseLine
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

func stateError(s *scriptScanner, c int) int {
	return scanError
}

func (s *scriptScanner) error(c int, context string) int {
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
