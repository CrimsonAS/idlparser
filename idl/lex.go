package idl

import (
	"fmt"
	"strings"
)

const lexDebug = false

// A TokenID represents a type of token in an IDL file.
type TokenID int

// Turn a TokenID into a string.
func (tok TokenID) String() string {
	val := ""
	switch tok {
	case TokenIdentifier:
		val = "identifier"
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
	case TokenOpenSquareBracket:
		val = "["
	case TokenCloseSquareBracket:
		val = "]"
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
	case TokenLessThan:
		val = "<"
	case TokenGreaterThan:
		val = ">"
	case TokenNamespace:
		val = "::"
	default:
		val = "(wtf)"
	}

	return fmt.Sprintf("%s(%d)", val, tok)
}

const (
	// TokenIdentifier represents any valid identifier
	TokenIdentifier = iota

	// TokenHash is a # character.
	TokenHash

	// TokenStringLiteral represents a quoted string
	TokenStringLiteral

	// TokenColon is a : character.
	TokenColon

	// TokenSemicolon is a ; character.
	TokenSemicolon

	// TokenOpenBrace is a { character.
	TokenOpenBrace

	// TokenCloseBrace is a } character.
	TokenCloseBrace

	// TokenOpenSquareBracket is a [ character.
	TokenOpenSquareBracket

	// TokenCloseSquareBracket is a ] character.
	TokenCloseSquareBracket

	// TokenOpenBracket is a ( character.
	TokenOpenBracket

	// TokenCloseBracket is a ) character.
	TokenCloseBracket

	// TokenEquals is a = character.
	TokenEquals

	// TokenEndLine is a \n character.
	TokenEndLine

	// TokenComma is a , character.
	TokenComma

	// TokenLessThan is a < character.
	TokenLessThan

	// TokenGreaterThan is a > character.
	TokenGreaterThan

	// TokenNamespace represents a namespace separator (::) used in types.
	TokenNamespace

	// TokenInvalid is a non-existent token used in error handling.
	TokenInvalid
)

// A Token represents an individal piece of an IDL file.
type Token struct {
	// ### these should be public
	// Represents the type of token, e.g. TokenWord
	ID TokenID

	// Represents the associated data of a token. For instance, TokenWord will
	// have a value containing the word that was lexed.
	Value string
}

// Turn a Token into a string
func (tok Token) String() string {
	if len(tok.Value) > 0 {
		return fmt.Sprintf(`%s("%s")`, tok.ID, tok.Value)
	}
	return fmt.Sprintf("%s", tok.ID)
}

// A lexer is used to lex a byte stream in IDL format into a series of
// tokens. The token series can then be interpreted directly, or parsed to
// ensure validity and become usable in a higher form.
type lexer struct {
	buf    []byte
	pos    int
	errors []error
	tokens []Token
}

// Add the given token to the stream
func (lb *lexer) pushToken(tok TokenID, val string) {
	if lexDebug {
		if len(val) > 0 {
			fmt.Printf("Lexed token %s val %s\n", tok, val)
		} else {
			fmt.Printf("Lexed token %s\n", tok)
		}
	}
	lb.tokens = append(lb.tokens, Token{tok, val})
}

func (lb *lexer) reportError(err error) {
	lb.errors = append(lb.errors, err)
}

func (lb *lexer) hasError() bool {
	return len(lb.errors) != 0
}

func (lb *lexer) cur() byte {
	return lb.buf[lb.pos]
}

func (lb *lexer) skipWhitespace() {
	for !lb.atEnd() && (lb.cur() == ' ' || lb.cur() == '\t') {
		lb.advance()
	}
}

func (lb *lexer) readUntilNot(delims []byte) ([]byte, error) {
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

func (lb *lexer) readUntilMany(delims []byte) ([]byte, error) {
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

func (lb *lexer) readUntil(delim byte) ([]byte, error) {
	return lb.readUntilMany([]byte{delim})
}

func (lb *lexer) next() byte {
	return lb.buf[lb.pos+1]
}

func (lb *lexer) atEnd() bool {
	return lb.pos >= len(lb.buf)
}

func (lb *lexer) advance() {
	lb.pos++
}

func (lb *lexer) rewind() {
	lb.pos--
}

func (lb *lexer) lexComment() {
	if lb.next() == '/' {
		lb.readUntil('\n')
	}
}

func (lb *lexer) lexStringLiteral() {
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

const (
	keywordModule    = "module"
	keywordTypedef   = "typedef"
	keywordStruct    = "struct"
	keywordConst     = "const"
	keywordEnum      = "enum"
	keywordInterface = "interface"
	keywordUnion     = "union"
	keywordIn        = "in"
	keywordOut       = "out"
	keywordInOut     = "inout"
	keywordSwitch    = "switch"
	keywordCase      = "case"
)

// ### this needs to be improved to read types properly.
// i.e. handle seqeunce<long, 10>.
var validInIdentifiers = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_")

func (lb *lexer) lexWord() {
	buf, err := lb.readUntilNot(validInIdentifiers)
	if err != nil {
		lb.reportError(fmt.Errorf("EOF on a word?"))
	}

	lb.pushToken(TokenIdentifier, string(buf))
}

// Lex a buffer of IDL data into tokens.
// Returns the lexed tokens, and any error encountered.
func Lex(d []byte) ([]Token, error) {
	lb := &lexer{
		buf: d,
		pos: 0,
	}

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
		case lb.cur() == '[':
			lb.pushToken(TokenOpenSquareBracket, "")
		case lb.cur() == ']':
			lb.pushToken(TokenCloseSquareBracket, "")
		case lb.cur() == '(':
			lb.pushToken(TokenOpenBracket, "")
		case lb.cur() == ')':
			lb.pushToken(TokenCloseBracket, "")
		case lb.cur() == ':':
			lb.advance()
			if lb.cur() == ':' {
				lb.pushToken(TokenNamespace, "")
			} else {
				lb.rewind()
				lb.pushToken(TokenColon, "")
			}
		case lb.cur() == ';':
			lb.pushToken(TokenSemicolon, "")
		case lb.cur() == '=':
			lb.pushToken(TokenEquals, "")
		case lb.cur() == '\n':
			lb.pushToken(TokenEndLine, "")
		case lb.cur() == ',':
			lb.pushToken(TokenComma, "")
		case lb.cur() == '<':
			lb.pushToken(TokenLessThan, "")
		case lb.cur() == '>':
			lb.pushToken(TokenGreaterThan, "")
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
