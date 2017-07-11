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
func (l *lexer) pushToken(tok TokenID, val string) {
	if lexDebug {
		if len(val) > 0 {
			fmt.Printf("Lexed token %s val %s\n", tok, val)
		} else {
			fmt.Printf("Lexed token %s\n", tok)
		}
	}
	l.tokens = append(l.tokens, Token{tok, val})
}

func (l *lexer) reportError(err error) {
	l.errors = append(l.errors, err)
}

func (l *lexer) hasError() bool {
	return len(l.errors) != 0
}

func (l *lexer) cur() byte {
	return l.buf[l.pos]
}

func (l *lexer) skipWhitespace() {
	for !l.atEnd() && (l.cur() == ' ' || l.cur() == '\t') {
		l.advance()
	}
}

func (l *lexer) readUntilNot(delims []byte) ([]byte, error) {
	buf := []byte{}
	found := true
	for !l.atEnd() && found {
		found = false
		for _, delim := range delims {
			if l.cur() == delim {
				found = true
				break
			}
		}

		if found {
			buf = append(buf, l.cur())
			l.advance()
		} else {
			l.rewind()
		}
	}

	if l.atEnd() {
		return nil, fmt.Errorf("didn't find %s, when I wanted it", delims)
	}

	return buf, nil
}

func (l *lexer) readUntilMany(delims []byte) ([]byte, error) {
	buf := []byte{}
	found := false
	for !l.atEnd() && !found {
		for _, delim := range delims {
			if l.cur() == delim {
				found = true
				break
			}
		}

		if !found {
			buf = append(buf, l.cur())
			l.advance()
		}
	}

	if l.atEnd() {
		return nil, fmt.Errorf("didn't find %s, when I wanted it", delims)
	}

	return buf, nil
}

func (l *lexer) readUntil(delim byte) ([]byte, error) {
	return l.readUntilMany([]byte{delim})
}

func (l *lexer) next() byte {
	return l.buf[l.pos+1]
}

func (l *lexer) atEnd() bool {
	return l.pos >= len(l.buf)
}

func (l *lexer) advance() {
	l.pos++
}

func (l *lexer) rewind() {
	l.pos--
}

func (l *lexer) lexComment() {
	if l.next() == '/' {
		l.readUntil('\n')
	}
}

func (l *lexer) lexStringLiteral() {
	if l.cur() != '"' {
		l.reportError(fmt.Errorf("expected: \", got: %c", l.cur()))
		return
	}

	// skip "
	l.advance()

	buf, err := l.readUntil('"')
	if err != nil {
		l.reportError(fmt.Errorf("unterminated string literal"))
	}

	l.pushToken(TokenStringLiteral, string(buf))
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

func (l *lexer) lexWord() {
	buf, err := l.readUntilNot(validInIdentifiers)
	if err != nil {
		l.reportError(fmt.Errorf("EOF on a word?"))
	}

	l.pushToken(TokenIdentifier, string(buf))
}

// Lex a buffer of IDL data into tokens.
// Returns the lexed tokens, and any error encountered.
func Lex(d []byte) ([]Token, error) {
	l := &lexer{
		buf: d,
		pos: 0,
	}

	for !l.atEnd() && !l.hasError() {
		l.skipWhitespace()
		if l.atEnd() {
			break
		}

		switch {
		case l.cur() == '/':
			l.lexComment()
		case l.cur() == '#':
			l.pushToken(TokenHash, "")
		case l.cur() == '"':
			l.lexStringLiteral()
		case l.cur() == '{':
			l.pushToken(TokenOpenBrace, "")
		case l.cur() == '}':
			l.pushToken(TokenCloseBrace, "")
		case l.cur() == '[':
			l.pushToken(TokenOpenSquareBracket, "")
		case l.cur() == ']':
			l.pushToken(TokenCloseSquareBracket, "")
		case l.cur() == '(':
			l.pushToken(TokenOpenBracket, "")
		case l.cur() == ')':
			l.pushToken(TokenCloseBracket, "")
		case l.cur() == ':':
			l.advance()
			if l.cur() == ':' {
				l.pushToken(TokenNamespace, "")
			} else {
				l.rewind()
				l.pushToken(TokenColon, "")
			}
		case l.cur() == ';':
			l.pushToken(TokenSemicolon, "")
		case l.cur() == '=':
			l.pushToken(TokenEquals, "")
		case l.cur() == '\n':
			l.pushToken(TokenEndLine, "")
		case l.cur() == ',':
			l.pushToken(TokenComma, "")
		case l.cur() == '<':
			l.pushToken(TokenLessThan, "")
		case l.cur() == '>':
			l.pushToken(TokenGreaterThan, "")
		case strings.IndexByte(string(validInIdentifiers), l.cur()) >= 0:
			l.lexWord()
		}

		l.advance()
	}

	if l.hasError() {
		return []Token{}, l.errors[0]
	}

	return l.tokens, nil
}
