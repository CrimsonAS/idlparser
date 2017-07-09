package idl

import (
	"fmt"
	"strings"
)

type tokenId int

func (tok tokenId) String() string {
	val := ""
	switch tok {
	case tokenWord:
		val = "word"
	case tokenHash:
		val = "#"
	case tokenStringLiteral:
		val = "quoted string"
	case tokenSemicolon:
		val = ";"
	case tokenOpenBrace:
		val = "{"
	case tokenCloseBrace:
		val = "}"
	case tokenOpenBracket:
		val = "("
	case tokenCloseBracket:
		val = ")"
	case tokenEquals:
		val = "="
	default:
		val = "(wtf)"
	}

	return fmt.Sprintf("%s(%d)", val, tok)
}

const (
	// Any bare word, could be a keyword like module or a struct name.
	tokenWord = 99

	// '#'
	tokenHash = iota

	// A quoted string
	tokenStringLiteral = iota

	// ;
	tokenSemicolon = iota

	// {
	tokenOpenBrace = iota

	// }
	tokenCloseBrace = iota

	// (
	tokenOpenBracket = iota

	// )
	tokenCloseBracket = iota

	// =
	tokenEquals = iota
)

type token struct {
	id    tokenId
	value string
}

type LexBuf struct {
	buf    []byte
	pos    int
	errors []error
	tokens []token
}

func NewLexBuf(d []byte) *LexBuf {
	lb := &LexBuf{
		buf: d,
		pos: 0,
	}
	return lb
}

const lexDebug = false

func (lb *LexBuf) pushToken(tok tokenId, val string) {
	if lexDebug {
		if len(val) > 0 {
			fmt.Printf("Lexed token %s val %s\n", tok, val)
		} else {
			fmt.Printf("Lexed token %s\n", tok)
		}
	}
	lb.tokens = append(lb.tokens, token{tok, val})
}

func (lb *LexBuf) reportError(err error) {
	lb.errors = append(lb.errors, err)
}

func (lb *LexBuf) hasError() bool {
	return len(lb.errors) != 0
}

func (lb *LexBuf) cur() byte {
	return lb.buf[lb.pos]
}

func (lb *LexBuf) skipWhitespace() {
	for !lb.atEnd() && (lb.cur() == ' ' || lb.cur() == '\t' || lb.cur() == '\n') {
		lb.advance()
	}
}

func (lb *LexBuf) readUntilNot(delims []byte) ([]byte, error) {
	buf := []byte{}
	found := true
	for !lb.atEnd() && found {
		found = false
		for _, delim := range delims {
			if lb.cur() == delim {
				found = true
				break
			}
		}

		if found {
			buf = append(buf, lb.cur())
			lb.advance()
		}
	}

	if lb.atEnd() {
		return nil, fmt.Errorf("didn't find %s, when I wanted it", delims)
	}

	return buf, nil
}

func (lb *LexBuf) readUntilMany(delims []byte) ([]byte, error) {
	buf := []byte{}
	found := false
	for !lb.atEnd() && !found {
		for _, delim := range delims {
			if lb.cur() == delim {
				found = true
				break
			}
		}

		if !found {
			buf = append(buf, lb.cur())
			lb.advance()
		}
	}

	if lb.atEnd() {
		return nil, fmt.Errorf("didn't find %s, when I wanted it", delims)
	}

	return buf, nil
}

func (lb *LexBuf) readUntil(delim byte) ([]byte, error) {
	return lb.readUntilMany([]byte{delim})
}

func (lb *LexBuf) next() byte {
	return lb.buf[lb.pos+1]
}

func (lb *LexBuf) atEnd() bool {
	return lb.pos >= len(lb.buf)
}

func (lb *LexBuf) advance() {
	lb.pos += 1
}

func (lb *LexBuf) lexComment() {
	if lb.next() == '/' {
		lb.readUntil('\n')
	}
}

func (lb *LexBuf) lexStringLiteral() {
	if lb.cur() != '"' {
		lb.reportError(fmt.Errorf("expected: \", got: %c", lb.cur()))
		return
	}

	// skip "
	lb.advance()

	buf, err := lb.readUntil('"')
	if err != nil {
		lb.reportError(fmt.Errorf("unterminated string literal"))
	}

	lb.pushToken(tokenStringLiteral, string(buf))
}

var validInIdentifiers = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_")

func (lb *LexBuf) lexWord() {
	buf, err := lb.readUntilNot(validInIdentifiers)
	if err != nil {
		lb.reportError(fmt.Errorf("EOF on a word?"))
	}

	lb.pushToken(tokenWord, string(buf))

}

func (lb *LexBuf) Lex() error {
	for !lb.atEnd() && !lb.hasError() {
		lb.skipWhitespace()
		if lb.atEnd() {
			break
		}

		switch {
		case lb.cur() == '/':
			lb.lexComment()
		case lb.cur() == '#':
			lb.pushToken(tokenHash, "")
		case lb.cur() == '"':
			lb.lexStringLiteral()
		case lb.cur() == '{':
			lb.pushToken(tokenOpenBrace, "")
		case lb.cur() == '}':
			lb.pushToken(tokenCloseBrace, "")
		case lb.cur() == '(':
			lb.pushToken(tokenOpenBracket, "")
		case lb.cur() == ')':
			lb.pushToken(tokenCloseBracket, "")
		case lb.cur() == ';':
			lb.pushToken(tokenSemicolon, "")
		case lb.cur() == '=':
			lb.pushToken(tokenEquals, "")
		case strings.IndexByte(string(validInIdentifiers), lb.cur()) >= 0:
			lb.lexWord()
		}

		lb.advance()
	}

	if lb.hasError() {
		return lb.errors[0]
	}

	return nil
}
