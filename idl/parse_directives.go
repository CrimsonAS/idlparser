package idl

import (
	"fmt"
)

// The entry point for directives.
func (pb *parser) parseTokenHash() {
	pb.advance() // skip #

	if pb.tok().Id != TokenIdentifier {
		pb.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	directive := pb.tok().Value

	switch directive {
	case "define":
		pb.parseDefineDirective()
	case "include":
		pb.parseIncludeDirective()
	default:
		pb.reportError(fmt.Errorf("unexpected directive: %s", directive))
	}
}

func (pb *parser) parseDefineDirective() {
	pb.advance()

	if pb.tok().Id != TokenIdentifier {
		pb.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	varName := pb.tok().Value
	pb.advanceAndDontSkipNewLines()

	if !pb.atEnd() && pb.tok().Id == TokenIdentifier {
		varValue := pb.tok().Value
		// Don't skip newlines so that:
		// #define FOO
		// Something
		// isn't treated as "#define FOO Something".
		pb.advanceAndDontSkipNewLines()
		fmt.Printf("Define: %s val %s\n", varName, varValue)
	} else {
		fmt.Printf("Define: %s no value\n", varName)
	}
}

func (pb *parser) parseIncludeDirective() {
	pb.advance()

	if pb.tok().Id != TokenStringLiteral {
		pb.reportError(fmt.Errorf("unexpected non-string-literal"))
		return
	}

	fileName := pb.tok().Value
	pb.advance()

	fmt.Printf("Included: %s\n", fileName)
}
