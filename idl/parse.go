package idl

import (
	"fmt"
)

const parseDebug = true

type context int32

const (
	// Outermost
	contextGlobal = iota

	// In a module
	contextModule = iota

	// In a struct
	contextStruct = iota
)

type ParseBuf struct {
	lb           *LexBuf
	contextStack []context
	ppos         int
	errors       []error
}

func (pb *ParseBuf) reportError(err error) {
	fmt.Printf("Got parse error: %s\n", err)
	pb.errors = append(pb.errors, err)
}

func (pb *ParseBuf) hasError() bool {
	return len(pb.errors) != 0
}

func NewParseBuf(lexBuf *LexBuf) *ParseBuf {
	pb := &ParseBuf{
		lb: lexBuf,
	}
	return pb
}

func (pb *ParseBuf) parseDefineDirective() {
	pb.advance()
	if pb.atEnd() {
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return
	}

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	varName := pb.tok().value
	pb.advance()

	if !pb.atEnd() && pb.tok().id == tokenWord {
		varValue := pb.tok().value
		pb.advance()
		fmt.Printf("Define: %s val %s\n", varName, varValue)
	} else {
		fmt.Printf("Define: %s no value\n", varName)
	}
}

func (pb *ParseBuf) parseIncludeDirective() {
	pb.advance()
	if pb.atEnd() {
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return
	}

	if pb.tok().id != tokenStringLiteral {
		pb.reportError(fmt.Errorf("unexpected non-string-literal"))
		return
	}

	fileName := pb.tok().value
	pb.advance()

	fmt.Printf("Included: %s\n", fileName)
}

// The entry point for directives.
func (pb *ParseBuf) parseTokenHash() {
	pb.advance() // skip #

	if pb.atEnd() {
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return
	}

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	directive := pb.tok().value

	switch directive {
	case "define":
		pb.parseDefineDirective()
	case "include":
		pb.parseIncludeDirective()
	default:
		pb.reportError(fmt.Errorf("unexpected directive: %s", directive))
	}
}

func (pb *ParseBuf) parseTokenWord() {
	word := pb.tok().value

	fmt.Printf("Got word %s\n", word)
	if pb.currentContext() == contextGlobal {
		if word != "module" {
			pb.reportError(fmt.Errorf("unexpected keyword in global context: %s", word))
			return
		}

		pb.pushContext(contextModule)
		pb.advance()

		if pb.atEnd() {
			pb.reportError(fmt.Errorf("unexpected EOF"))
			return
		}

		if pb.tok().id != tokenWord {
			pb.reportError(fmt.Errorf("expected module name"))
			return
		}

		fmt.Printf("Opened module %s\n", pb.tok().value)
	}
}

func (pb *ParseBuf) atEnd() bool {
	// ### right?
	return pb.ppos >= len(pb.lb.tokens)-1
}

func (pb *ParseBuf) tok() token {
	return pb.lb.tokens[pb.ppos]
}

func (pb *ParseBuf) advance() {
	if parseDebug {
		fmt.Printf("Advancing, ppos was %d\n", pb.ppos)
	}
	pb.ppos += 1
}

func (pb *ParseBuf) Parse() {
	pb.pushContext(contextGlobal)

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
		default:
			pb.advance()
		}
	}

	pb.popContext()
	if len(pb.contextStack) > 0 {
		panic("too many contexts")
	}
}

func (pb *ParseBuf) pushContext(ctx context) {
	pb.contextStack = append(pb.contextStack, ctx)
}

func (pb *ParseBuf) popContext() {
	pb.contextStack = pb.contextStack[:len(pb.contextStack)-1]
}

func (pb *ParseBuf) currentContext() context {
	return pb.contextStack[len(pb.contextStack)-1]
}
