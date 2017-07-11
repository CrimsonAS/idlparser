package idl

import (
	"fmt"
)

const parseDebug = false

// A context id is used to drive the internal state machine. It is not needed
// outside the parser.
type contextID int32

func (ctx contextID) String() string {
	switch ctx {
	case contextGlobal:
		return "global"
	case contextModule:
		return "module"
	case contextStruct:
		return "struct"
	case contextEnum:
		return "enum"
	case contextInterface:
		return "interface"
	case contextUnion:
		return "union"
	}

	return "(wtf)"
}

type context struct {
	id    contextID
	value string
}

const (
	// Outermost
	contextGlobal = iota

	// In a module
	contextModule

	// In a struct
	contextStruct

	// In an enum
	contextEnum

	// In an interface
	contextInterface

	// In a union
	contextUnion
)

// A parser parses IDL into an AST representation. It consumes a series of lexed
// tokens.
type parser struct {
	tokens       []Token
	contextStack []context
	ppos         int
	errors       []error
	isEOF        bool

	// current module being populated
	currentModule *Module

	// ### these might belong on the AST instead? perhaps not, since they are
	// only useful during setup...
	currentEnum   *Enum
	currentStruct *Struct
	currentIface  *Interface

	// root module that everything belongs in
	rootModule *Module
}

// Report an error during parsing. Further parsing will be aborted.
func (p *parser) reportError(err error) {
	// Ignore all errors after EOF, as they are likely bogus (due to our
	// returning a silly token in that case to avoid crashes).
	if !p.isEOF {
		fmt.Printf("Got parse error: %s\n", err)
		p.errors = append(p.errors, err)
	}
}

// Does the parsing have an error already?
func (p *parser) hasError() bool {
	return len(p.errors) != 0
}

// Small helper to read a type name. A type name is a bit "special" since it
// might be one word ("int"), or multiple ("unsigned int").
func (p *parser) parseType() Type {
	t := Type{}

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected type name"))
		return t
	}

	t.Name = p.tok().Value
	p.advance()

	// sequence<foo>, string<foo>
	// ### this should come after the namespace check
	if p.tok().ID == TokenLessThan {
		p.advance()
		t.TemplateParameters = append(t.TemplateParameters, p.parseType())

		for p.tok().ID == TokenComma {
			p.advance()
			t.TemplateParameters = append(t.TemplateParameters, p.parseType())
		}

		if p.tok().ID != TokenGreaterThan {
			fmt.Errorf("expected: >")
			return t
		}

		p.advance()
	}

	if t.Name == "unsigned" {
		// consume an additional word
		if p.tok().ID != TokenIdentifier {
			p.reportError(fmt.Errorf("expected numeric type"))
			return t
		}

		t.Name += " " + p.tok().Value
		p.advance()
	} else if t.Name == "long" {
		// "long long"
		if p.tok().ID == TokenIdentifier && p.tok().Value == "long" {
			t.Name += " " + p.tok().Value
			p.advance()
		}
	}

	if p.tok().ID == TokenNamespace {
		// Foo::Bar
		p.advance()
		t.Name += "::" + p.tok().Value
		p.advance()
	}

	return t
}

func (p *parser) parseIdentifier() string {
	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected identifier"))
		return ""
	}

	identifierName := p.tok().Value
	p.advance()

	if p.tok().ID == TokenNamespace {
		// Foo::Bar
		p.advance()
		if p.tok().ID != TokenIdentifier {
			p.reportError(fmt.Errorf("expected type name in namespace"))
			return ""
		}

		identifierName += "::" + p.tok().Value
		p.advance()
	}

	if p.tok().ID == TokenOpenSquareBracket {
		// value[3]
		p.advance()
		identifierName += "["

		if p.tok().ID != TokenIdentifier {
			p.reportError(fmt.Errorf("expected quantity"))
			return ""
		}

		identifierName += p.tok().Value
		p.advance()

		if p.tok().ID != TokenCloseSquareBracket {
			p.reportError(fmt.Errorf("expected close bracket"))
			return ""
		}

		p.advance()
		identifierName += "]"
	}

	return identifierName
}

func (p *parser) parseValue() string {
	if p.tok().ID != TokenIdentifier &&
		p.tok().ID != TokenLessThan &&
		p.tok().ID != TokenStringLiteral {
		p.reportError(fmt.Errorf("expected value"))
		return ""
	}

	val := ""
	isString := false
	if p.tok().ID == TokenStringLiteral {
		isString = true
		val = "\""
	}

	for p.tok().ID == TokenIdentifier ||
		p.tok().ID == TokenLessThan ||
		p.tok().ID == TokenStringLiteral {
		val += p.tok().Value
		p.advance()
	}

	if isString {
		val += "\""
	}

	return val
}

// Parse a regular word. It might be a keyword (like 'struct' or 'module', or it
// might be a type name (in struct or interface members).
func (p *parser) parseTokenWord() {
	word := p.tok().Value

	switch p.currentContext().id {
	case contextGlobal:
		fallthrough
	case contextModule:
		// Regular contexts only allow a certain set of keywords.
		switch word {
		case keywordModule:
			p.parseModule()
		case keywordTypedef:
			p.parseTypedef()
		case keywordStruct:
			p.parseStruct()
		case keywordConst:
			p.parseConst()
		case keywordEnum:
			p.parseEnum()
		case keywordInterface:
			p.parseInterface()
		case keywordUnion:
			p.parseUnion()
		default:
			p.reportError(fmt.Errorf("unexpected keyword in global/module context: %s", word))
			return
		}

	// For other contexts, they are supposed to validate the contents
	// themselves.
	case contextStruct:
		p.parseStructMember()
	case contextEnum:
		p.parseEnumMember()
	case contextInterface:
		p.parseInterfaceMember()
	case contextUnion:
		p.parseUnionMember()
	default:
		panic("unhandled context")
	}
}

// Is the parser at the end of the token stream?
func (p *parser) atEnd() bool {
	// ### right?
	return p.ppos >= len(p.tokens)-1
}

// Return the next token for parsing
func (p *parser) peekTok() Token {
	if p.atEnd() || p.ppos+1 >= len(p.tokens)-1 {
		if parseDebug {
			fmt.Printf("Peeking ahead invalid!\n")
		}
		return Token{TokenInvalid, ""}
	}
	if parseDebug {
		fmt.Printf("Peeking ahead ppos %d is %s\n", p.ppos, p.tokens[p.ppos+1])
	}
	return p.tokens[p.ppos+1]
}

// Return the current token under parsing
func (p *parser) tok() Token {
	if p.atEnd() {
		p.isEOF = true
		p.reportError(fmt.Errorf("unexpected EOF"))
		return Token{TokenInvalid, ""}
	}
	return p.tokens[p.ppos]
}

// Advance the parse stream to the next non-newline token.
func (p *parser) advance() {
	for {
		p.advanceAndDontSkipNewLines()

		// Skip all whitespace tokens.
		if p.tok().ID != TokenEndLine {
			break
		}
	}
}

// Advance the parse stream one position
func (p *parser) advanceAndDontSkipNewLines() {
	if parseDebug {
		fmt.Printf("Advancing, ppos was %d/%d, old token %s new token %s\n", p.ppos, len(p.tokens), p.tokens[p.ppos], p.tokens[p.ppos+1])
	}
	p.ppos++
}

// Parse a series of tokens, and return an AST representing the IDL's content.
func Parse(toks []Token) (Module, error) {
	p := &parser{
		tokens:        toks,
		isEOF:         false,
		currentModule: &Module{},
	}
	p.rootModule = p.currentModule
	p.pushContext(contextGlobal, "")

	for !p.atEnd() && !p.hasError() {
		tok := p.tok()
		if parseDebug {
			if len(tok.Value) > 0 {
				fmt.Printf("ppos %d Parsing token %s\n", p.ppos, tok)
			}
		}

		switch tok.ID {
		case TokenHash:
			p.parseTokenHash()
		case TokenIdentifier:
			p.parseTokenWord()
		case TokenCloseBrace:
			p.popContext()
			p.advance()
		default:
			p.advance()
		}
	}

	p.popContext()
	if p.hasError() {
		return Module{}, p.errors[0]
	}

	if len(p.contextStack) > 0 {
		panic("too many contexts")
	}
	return *p.rootModule, nil
}

func (p *parser) pushContext(ctx contextID, val string) {
	if parseDebug {
		fmt.Printf("Opened context: %s (%s)\n", ctx, val)
	}

	switch ctx {
	case contextInterface:
		e := Interface{Name: val}
		p.currentModule.Interfaces = append(p.currentModule.Interfaces, e)
		p.currentIface = &p.currentModule.Interfaces[len(p.currentModule.Interfaces)-1]
	case contextStruct:
		e := Struct{Name: val}
		p.currentModule.Structs = append(p.currentModule.Structs, e)
		p.currentStruct = &p.currentModule.Structs[len(p.currentModule.Structs)-1]
	case contextEnum:
		e := Enum{Name: val}
		p.currentModule.Enums = append(p.currentModule.Enums, e)
		p.currentEnum = &p.currentModule.Enums[len(p.currentModule.Enums)-1]
	case contextModule:
		m := Module{
			Name:   val,
			parent: p.currentModule,
		}
		p.currentModule.Modules = append(p.currentModule.Modules, m)
		p.currentModule = &p.currentModule.Modules[len(p.currentModule.Modules)-1]
	}

	p.contextStack = append(p.contextStack, context{ctx, val})
}

func (p *parser) popContext() {
	cctx := p.currentContext()

	switch cctx.id {
	case contextInterface:
		p.currentIface = nil
	case contextStruct:
		p.currentStruct = nil
	case contextEnum:
		p.currentEnum = nil
	case contextModule:
		if p.currentModule.parent != nil {
			p.currentModule = p.currentModule.parent
		}
	}

	if parseDebug {
		fmt.Printf("Closed context: %s (%s)\n", cctx.id, cctx.value)
	}
	p.contextStack = p.contextStack[:len(p.contextStack)-1]
}

func (p *parser) currentContext() context {
	return p.contextStack[len(p.contextStack)-1]
}
