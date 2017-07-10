package idl

import (
	"fmt"
)

const parseDebug = false

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
)

// A ParseBuf is a parser. It consumes a series of lexed tokens to understand
// the IDL file.
type ParseBuf struct {
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
func (pb *ParseBuf) reportError(err error) {
	// Ignore all errors after EOF, as they are likely bogus (due to our
	// returning a silly token in that case to avoid crashes).
	if !pb.isEof {
		fmt.Printf("Got parse error: %s\n", err)
		pb.errors = append(pb.errors, err)
	}
}

// Does the parsing have an error already?
func (pb *ParseBuf) hasError() bool {
	return len(pb.errors) != 0
}

// Create a ParseBuf from a LexBuf's tokens.
func NewParseBuf(toks []Token) *ParseBuf {
	pb := &ParseBuf{
		tokens:        toks,
		isEof:         false,
		currentModule: &Module{},
	}
	pb.rootModule = pb.currentModule
	return pb
}

// Small helper to read a type name. A type name is a bit "special" since it
// might be one word ("int"), or multiple ("unsigned int").
func (pb *ParseBuf) parseType() string {
	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected constant type"))
		return ""
	}

	constType := pb.tok().Value
	pb.advance()

	if constType == "unsigned" {
		// consume an additional word
		if pb.tok().Id != TokenWord {
			pb.reportError(fmt.Errorf("expected numeric type"))
			return ""
		}

		constType += " " + pb.tok().Value
		pb.advance()
	}

	return constType
}

// Parse a regular word. It might be a keyword (like 'struct' or 'module', or it
// might be a type name (in struct or interface members).
func (pb *ParseBuf) parseTokenWord() {
	word := pb.tok().Value

	switch pb.currentContext().id {
	case contextGlobal:
		fallthrough
	case contextModule:
		// Regular contexts only allow a certain set of keywords.
		switch word {
		case "module":
			pb.parseModule()
		case "typedef":
			pb.parseTypedef()
		case "struct":
			pb.parseStruct()
		case "const":
			pb.parseConst()
		case "enum":
			pb.parseEnum()
		case "interface":
			pb.parseInterface()
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

// Is the ParseBuf at the end of the token stream?
func (pb *ParseBuf) atEnd() bool {
	// ### right?
	return pb.ppos >= len(pb.tokens)-1
}

// Return the current token under parsing
func (pb *ParseBuf) tok() Token {
	if pb.atEnd() {
		pb.isEof = true
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return Token{TokenInvalid, ""}
	}
	return pb.tokens[pb.ppos]
}

// Advance the parse stream to the next non-newline token.
func (pb *ParseBuf) advance() {
	for {
		pb.advanceAndDontSkipNewLines()

		// Skip all whitespace tokens.
		if pb.tok().Id != TokenEndLine {
			break
		}
	}
}

// Advance the parse stream one position
func (pb *ParseBuf) advanceAndDontSkipNewLines() {
	if parseDebug {
		fmt.Printf("Advancing, ppos was %d, old token %s new token %s\n", pb.ppos, pb.tokens[pb.ppos], pb.tokens[pb.ppos+1])
	}
	pb.ppos += 1
}

// Initiate the parsing process
func (pb *ParseBuf) Parse() (Module, error) {
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
		case TokenWord:
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

func (pb *ParseBuf) pushContext(ctx contextId, val string) {
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

func (pb *ParseBuf) popContext() {
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

func (pb *ParseBuf) currentContext() context {
	return pb.contextStack[len(pb.contextStack)-1]
}
