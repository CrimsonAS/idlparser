package idl

import (
	"fmt"
)

// The entry point for directives.
func (p *parser) parseTokenHash() {
	p.advance() // skip #

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	directive := p.tok().Value

	switch directive {
	case "define":
		p.parseDefineDirective()
	case "include":
		p.parseIncludeDirective()
	default:
		p.reportError(fmt.Errorf("unexpected directive: %s", directive))
	}
}

func (p *parser) parseDefineDirective() {
	p.advance()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	varName := p.tok().Value
	p.advanceAndDontSkipNewLines()

	if !p.atEnd() && p.tok().ID == TokenIdentifier {
		varValue := p.tok().Value
		// Don't skip newlines so that:
		// #define FOO
		// Something
		// isn't treated as "#define FOO Something".
		p.advanceAndDontSkipNewLines()
		fmt.Printf("Define: %s val %s\n", varName, varValue)
	} else {
		fmt.Printf("Define: %s no value\n", varName)
	}
}

func (p *parser) parseIncludeDirective() {
	p.advance()

	if p.tok().ID != TokenStringLiteral {
		p.reportError(fmt.Errorf("unexpected non-string-literal"))
		return
	}

	fileName := p.tok().Value
	p.advance()

	fmt.Printf("Included: %s\n", fileName)
}
