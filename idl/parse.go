package idl

import (
	"fmt"
)

const parseDebug = true

// A context id is used to drive the internal state machine. It is not needed
// outside the parser.
type contextId int32

func (ctx contextId) String() string {
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
	id    contextId
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
	isEof        bool

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
func (pb *parser) reportError(err error) {
	// Ignore all errors after EOF, as they are likely bogus (due to our
	// returning a silly token in that case to avoid crashes).
	if !pb.isEof {
		fmt.Printf("Got parse error: %s\n", err)
		pb.errors = append(pb.errors, err)
	}
}

// Does the parsing have an error already?
func (pb *parser) hasError() bool {
	return len(pb.errors) != 0
}

// Small helper to read a type name. A type name is a bit "special" since it
// might be one word ("int"), or multiple ("unsigned int").
func (pb *parser) parseType() string {
	if pb.tok().Id != TokenIdentifier && (pb.tok().Id != TokenKeyword) {
		pb.reportError(fmt.Errorf("expected type name"))
		return ""
	}

	typeName := pb.tok().Value
	pb.advance()

	// sequence<foo>, string<foo>
	if pb.tok().Id == TokenLessThan {
		pb.advance()
		typeName += "<" + pb.parseType()

		for pb.tok().Id == TokenComma {
			typeName += ", " + pb.parseType()
		}

		if pb.tok().Id != TokenGreaterThan {
			fmt.Errorf("expected: >")
			return ""
		}

		typeName += ">"
		pb.advance()
	}

	if typeName == "unsigned" {
		// consume an additional word
		if pb.tok().Id != TokenIdentifier {
			pb.reportError(fmt.Errorf("expected numeric type"))
			return ""
		}

		typeName += " " + pb.tok().Value
		pb.advance()
	} else if typeName == "long" {
		// "long long"
		if pb.tok().Id == TokenIdentifier && pb.tok().Value == "long" {
			typeName += " " + pb.tok().Value
			pb.advance()
		}
	}

	if pb.tok().Id == TokenNamespace {
		// Foo::Bar
		pb.advance()
		typeName += "::" + pb.tok().Value
		pb.advance()
	}

	return typeName
}

func (p *parser) parseIdentifier() string {
	if p.tok().Id != TokenIdentifier {
		p.reportError(fmt.Errorf("expected identifier"))
		return ""
	}

	identifierName := p.tok().Value
	p.advance()

	if p.tok().Id == TokenNamespace {
		// Foo::Bar
		p.advance()
		if p.tok().Id != TokenIdentifier {
			p.reportError(fmt.Errorf("expected type name in namespace"))
			return ""
		}

		identifierName += "::" + p.tok().Value
		p.advance()
	}

	if p.tok().Id == TokenOpenSquareBracket {
		// value[3]
		p.advance()
		identifierName += "["

		if p.tok().Id != TokenIdentifier {
			p.reportError(fmt.Errorf("expected quantity"))
			return ""
		}

		identifierName += p.tok().Value
		p.advance()

		if p.tok().Id != TokenCloseSquareBracket {
			p.reportError(fmt.Errorf("expected close bracket"))
			return ""
		}

		p.advance()
	}

	return identifierName
}

func (p *parser) parseValue() string {
	if p.tok().Id != TokenIdentifier &&
		p.tok().Id != TokenLessThan &&
		p.tok().Id != TokenStringLiteral {
		p.reportError(fmt.Errorf("expected value"))
		return ""
	}

	val := ""
	for p.tok().Id == TokenIdentifier ||
		p.tok().Id == TokenLessThan ||
		p.tok().Id == TokenStringLiteral {
		val += p.tok().Value
		p.advance()
	}

	return val
}

// Parse a regular word. It might be a keyword (like 'struct' or 'module', or it
// might be a type name (in struct or interface members).
func (pb *parser) parseTokenWord() {
	word := pb.tok().Value

	switch pb.currentContext().id {
	case contextGlobal:
		fallthrough
	case contextModule:
		// Regular contexts only allow a certain set of keywords.
		switch word {
		case keywordModule:
			pb.parseModule()
		case keywordTypedef:
			pb.parseTypedef()
		case keywordStruct:
			pb.parseStruct()
		case keywordConst:
			pb.parseConst()
		case keywordEnum:
			pb.parseEnum()
		case keywordInterface:
			pb.parseInterface()
		case keywordUnion:
			pb.parseUnion()
		default:
			pb.reportError(fmt.Errorf("unexpected keyword in global/module context: %s", word))
			return
		}

	// For other contexts, they are supposed to validate the contents
	// themselves.
	case contextStruct:
		pb.parseStructMember()
	case contextEnum:
		pb.parseEnumMember()
	case contextInterface:
		pb.parseInterfaceMember()
	}
}

// Is the parser at the end of the token stream?
func (pb *parser) atEnd() bool {
	// ### right?
	return pb.ppos >= len(pb.tokens)-1
}

// Return the next token for parsing
func (pb *parser) peekTok() Token {
	if pb.atEnd() || pb.ppos+1 >= len(pb.tokens)-1 {
		if parseDebug {
			fmt.Printf("Peeking ahead invalid!\n")
		}
		return Token{TokenInvalid, ""}
	}
	if parseDebug {
		fmt.Printf("Peeking ahead ppos %d is %s\n", pb.ppos, pb.tokens[pb.ppos+1])
	}
	return pb.tokens[pb.ppos+1]
}

// Return the current token under parsing
func (pb *parser) tok() Token {
	if pb.atEnd() {
		pb.isEof = true
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return Token{TokenInvalid, ""}
	}
	return pb.tokens[pb.ppos]
}

// Advance the parse stream to the next non-newline token.
func (pb *parser) advance() {
	for {
		pb.advanceAndDontSkipNewLines()

		// Skip all whitespace tokens.
		if pb.tok().Id != TokenEndLine {
			break
		}
	}
}

// Advance the parse stream one position
func (pb *parser) advanceAndDontSkipNewLines() {
	if parseDebug {
		fmt.Printf("Advancing, ppos was %d/%d, old token %s new token %s\n", pb.ppos, len(pb.tokens), pb.tokens[pb.ppos], pb.tokens[pb.ppos+1])
	}
	pb.ppos += 1
}

// Parse a series of tokens, and return an AST representing the IDL's content.
func Parse(toks []Token) (Module, error) {
	pb := &parser{
		tokens:        toks,
		isEof:         false,
		currentModule: &Module{},
	}
	pb.rootModule = pb.currentModule
	pb.pushContext(contextGlobal, "")

	for !pb.atEnd() && !pb.hasError() {
		tok := pb.tok()
		if parseDebug {
			if len(tok.Value) > 0 {
				fmt.Printf("ppos %d Parsing token %s val %s\n", pb.ppos, tok.Id, tok.Value)
			} else {
				fmt.Printf("ppos %d Parsing token %s\n", pb.ppos, tok.Id)
			}
		}

		switch tok.Id {
		case TokenHash:
			pb.parseTokenHash()
		case TokenKeyword:
			fallthrough
		case TokenIdentifier:
			pb.parseTokenWord()
		case TokenCloseBrace:
			pb.popContext()
			pb.advance()
		default:
			pb.advance()
		}
	}

	pb.popContext()
	if pb.hasError() {
		return Module{}, pb.errors[0]
	}

	if len(pb.contextStack) > 0 {
		panic("too many contexts")
	}
	return *pb.rootModule, nil
}

func (pb *parser) pushContext(ctx contextId, val string) {
	if parseDebug {
		fmt.Printf("Opened context: %s (%s)\n", ctx, val)
	}

	switch ctx {
	case contextInterface:
		e := Interface{Name: val}
		pb.currentModule.Interfaces = append(pb.currentModule.Interfaces, e)
		pb.currentIface = &pb.currentModule.Interfaces[len(pb.currentModule.Interfaces)-1]
	case contextStruct:
		e := Struct{Name: val}
		pb.currentModule.Structs = append(pb.currentModule.Structs, e)
		pb.currentStruct = &pb.currentModule.Structs[len(pb.currentModule.Structs)-1]
	case contextEnum:
		e := Enum{Name: val}
		pb.currentModule.Enums = append(pb.currentModule.Enums, e)
		pb.currentEnum = &pb.currentModule.Enums[len(pb.currentModule.Enums)-1]
	case contextModule:
		m := Module{
			Name:   val,
			parent: pb.currentModule,
		}
		pb.currentModule.Modules = append(pb.currentModule.Modules, m)
		pb.currentModule = &pb.currentModule.Modules[len(pb.currentModule.Modules)-1]
	}

	pb.contextStack = append(pb.contextStack, context{ctx, val})
}

func (pb *parser) popContext() {
	cctx := pb.currentContext()

	switch cctx.id {
	case contextInterface:
		pb.currentIface = nil
	case contextStruct:
		pb.currentStruct = nil
	case contextEnum:
		pb.currentEnum = nil
	case contextModule:
		if pb.currentModule.parent != nil {
			pb.currentModule = pb.currentModule.parent
		}
	}

	if parseDebug {
		fmt.Printf("Closed context: %s (%s)\n", cctx.id, cctx.value)
	}
	pb.contextStack = pb.contextStack[:len(pb.contextStack)-1]
}

func (pb *parser) currentContext() context {
	return pb.contextStack[len(pb.contextStack)-1]
}
