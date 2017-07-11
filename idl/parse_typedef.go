package idl

import (
	"fmt"
)

func (p *parser) parseTypedef() {
	p.advance()

	fromName := p.parseType()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected to name"))
		return
	}

	toName := p.parseIdentifier()

	if p.tok().ID != TokenSemicolon {
		p.reportError(fmt.Errorf("expected semicolon, got: %s", p.tok().ID))
		return
	}

	p.advance()
	p.currentModule.TypeDefs = append(p.currentModule.TypeDefs, TypeDef{
		Name: toName,
		Type: fromName,
	})
	if parseDebug {
		fmt.Printf("Typedef: %s to %s\n", fromName, toName)
	}
}
