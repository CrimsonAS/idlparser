package idl

import (
	"fmt"
	"strings"
)

const lexDebug = false

// A TokenId represents a type of token in an IDL file.
type TokenId int

// Turn a TokenId into a string.
func (tok TokenId) String() string {
	val := ""
	switch tok {
	case TokenWord:
		val = "word"
	case TokenHash:
		val = "#"
	case TokenStringLiteral:
		val = "quoted string"
	case TokenColon:
		val = ":"
	case TokenSemicolon:
		val = ";"
	case TokenOpenBrace:
		val = "{"
	case TokenCloseBrace:
		val = "}"
	case TokenOpenBracket:
		val = "("
	case TokenCloseBracket:
		val = ")"
	case TokenEquals:
		val = "="
	case TokenEndLine:
		val = "\\n"
	case TokenComma:
		val = ","
	case TokenInvalid:
		val = "(invalid)"
	default:
		val = "(wtf)"
	}

	return fmt.Sprintf("%s(%d)", val, tok)
}

const (
	// Any bare word, could be a keyword like module or a struct name.
	TokenWord = iota

	// '#'
	TokenHash

	// A quoted string
	TokenStringLiteral

	// :
	TokenColon

	// ;
	TokenSemicolon

	// {
	TokenOpenBrace

	// }
	TokenCloseBrace

	// (
	TokenOpenBracket

	// )
	TokenCloseBracket

	// =
	TokenEquals

	// \n
	TokenEndLine

	// ,
	TokenComma

	// Used for error handling
	TokenInvalid
)

// A Token represents an individal piece of an IDL file.
type Token struct {
	// ### these should be public
	// Represents the type of token, e.g. TokenWord
	Id TokenId

	// Represents the associated data of a token. For instance, TokenWord will
	// have a value containing the word that was lexed.
	Value string
}

// Turn a Token into a string
func (tok Token) String() string {
	return fmt.Sprintf("%s(%s)", tok.Id, tok.Value)
}

// A LexBuf is a Lexer to lex a byte stream in IDL format into a series of
// tokens. The token series can then be interpreted directly, or parsed to
// ensure validity and become usable in a higher form.
type LexBuf struct {
	buf    []byte
	pos    int
	errors []error
	tokens []Token
}

// Create a lexer operating on the given data.
func NewLexBuf(d []byte) *LexBuf {
	lb := &LexBuf{
		buf: d,
		pos: 0,
	}
	return lb
}

// Add the given token to the stream
func (lb *LexBuf) pushToken(tok TokenId, val string) {
	if lexDebug {
		if len(val) > 0 {
			fmt.Printf("Lexed token %s val %s\n", tok, val)
		} else {
			fmt.Printf("Lexed token %s\n", tok)
		}
	}
	lb.tokens = append(lb.tokens, Token{tok, val})
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
	for !lb.atEnd() && (lb.cur() == ' ' || lb.cur() == '\t') {
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
		} else {
			lb.rewind()
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

func (lb *LexBuf) rewind() {
	lb.pos -= 1
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

	lb.pushToken(TokenStringLiteral, string(buf))
}

var validInIdentifiers = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_<>[]:")

func (lb *LexBuf) lexWord() {
	buf, err := lb.readUntilNot(validInIdentifiers)
	if err != nil {
		lb.reportError(fmt.Errorf("EOF on a word?"))
	}

	lb.pushToken(TokenWord, string(buf))

}

// Run the lex process.
func (lb *LexBuf) Lex() ([]Token, error) {
	for !lb.atEnd() && !lb.hasError() {
		lb.skipWhitespace()
		if lb.atEnd() {
			break
		}

		switch {
		case lb.cur() == '/':
			lb.lexComment()
		case lb.cur() == '#':
			lb.pushToken(TokenHash, "")
		case lb.cur() == '"':
			lb.lexStringLiteral()
		case lb.cur() == '{':
			lb.pushToken(TokenOpenBrace, "")
		case lb.cur() == '}':
			lb.pushToken(TokenCloseBrace, "")
		case lb.cur() == '(':
			lb.pushToken(TokenOpenBracket, "")
		case lb.cur() == ')':
			lb.pushToken(TokenCloseBracket, "")
		case lb.cur() == ':':
			lb.pushToken(TokenColon, "")
		case lb.cur() == ';':
			lb.pushToken(TokenSemicolon, "")
		case lb.cur() == '=':
			lb.pushToken(TokenEquals, "")
		case lb.cur() == '\n':
			lb.pushToken(TokenEndLine, "")
		case lb.cur() == ',':
			lb.pushToken(TokenComma, "")
		case strings.IndexByte(string(validInIdentifiers), lb.cur()) >= 0:
			lb.lexWord()
		}

		lb.advance()
	}

	if lb.hasError() {
		return []Token{}, lb.errors[0]
	}

	return lb.tokens, nil
}
