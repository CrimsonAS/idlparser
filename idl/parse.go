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
	contextModule = iota

	// In a struct
	contextStruct = iota

	// In an enum
	contextEnum = iota

	// In an interface
	contextInterface = iota
)

// A ParseBuf is a parser. It consumes a series of lexed tokens to understand
// the IDL file.
type ParseBuf struct {
	lb           *LexBuf
	contextStack []context
	ppos         int
	errors       []error
	isEof        bool
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
func NewParseBuf(lexBuf *LexBuf) *ParseBuf {
	pb := &ParseBuf{
		lb:    lexBuf,
		isEof: false,
	}
	return pb
}

// Small helper to read a type name. A type name is a bit "special" since it
// might be one word ("int"), or multiple ("unsigned int").
func (pb *ParseBuf) parseType() string {
	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected constant type"))
		return ""
	}

	constType := pb.tok().value
	pb.advance()

	if constType == "unsigned" {
		// consume an additional word
		if pb.tok().id != tokenWord {
			pb.reportError(fmt.Errorf("expected numeric type"))
			return ""
		}

		constType += " " + pb.tok().value
		pb.advance()
	}

	return constType
}

// Parse a regular word. It might be a keyword (like 'struct' or 'module', or it
// might be a type name (in struct or interface members).
func (pb *ParseBuf) parseTokenWord() {
	word := pb.tok().value

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
	return pb.ppos >= len(pb.lb.tokens)-1
}

// Return the current token under parsing
func (pb *ParseBuf) tok() token {
	if pb.atEnd() {
		pb.isEof = true
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return token{tokenEndLine, ""}
	}
	return pb.lb.tokens[pb.ppos]
}

// Advance the parse stream
func (pb *ParseBuf) advance() {
	if parseDebug {
		fmt.Printf("Advancing, ppos was %d, old token %s new token %s\n", pb.ppos, pb.lb.tokens[pb.ppos], pb.lb.tokens[pb.ppos+1])
	}
	pb.ppos += 1
}

// Initiate the parsing process
func (pb *ParseBuf) Parse() {
	pb.pushContext(contextGlobal, "")

	for !pb.atEnd() && !pb.hasError() {
		tok := pb.tok()
		if parseDebug {
			if len(tok.value) > 0 {
				fmt.Printf("ppos %d Parsing token %s val %s\n", pb.ppos, tok.id, tok.value)
			} else {
				fmt.Printf("ppos %d Parsing token %s\n", pb.ppos, tok.id)
			}
		}

		switch tok.id {
		case tokenHash:
			pb.parseTokenHash()
		case tokenWord:
			pb.parseTokenWord()
		case tokenCloseBrace:
			pb.popContext()
			pb.advance()
		default:
			pb.advance()
		}
	}

	pb.popContext()
	if len(pb.contextStack) > 0 {
		panic("too many contexts")
	}
}

func (pb *ParseBuf) pushContext(ctx contextId, val string) {
	fmt.Printf("Opened context: %s (%s)\n", ctx, val)
	pb.contextStack = append(pb.contextStack, context{ctx, val})
}

func (pb *ParseBuf) popContext() {
	cctx := pb.currentContext()
	fmt.Printf("Closed context: %s (%s)\n", cctx.id, cctx.value)
	pb.contextStack = pb.contextStack[:len(pb.contextStack)-1]
}

func (pb *ParseBuf) currentContext() context {
	return pb.contextStack[len(pb.contextStack)-1]
}
