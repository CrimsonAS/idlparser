package idl

import (
	"fmt"
)

// The entry point for directives.
func (pb *ParseBuf) parseTokenHash() {
	pb.advance() // skip #

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

func (pb *ParseBuf) parseDefineDirective() {
	pb.advance()

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

	if pb.tok().id != tokenStringLiteral {
		pb.reportError(fmt.Errorf("unexpected non-string-literal"))
		return
	}

	fileName := pb.tok().value
	pb.advance()

	fmt.Printf("Included: %s\n", fileName)
}
