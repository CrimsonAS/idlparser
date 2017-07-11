package idl

import (
	"fmt"
)

func (p *parser) parseModule() {
	p.advance()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected module name"))
		return
	}

	moduleName := p.parseIdentifier()

	if p.tok().ID != TokenOpenBrace {
		p.reportError(fmt.Errorf("expected module contents"))
		return
	}

	p.advance()
	p.pushContext(contextModule, moduleName)
}
