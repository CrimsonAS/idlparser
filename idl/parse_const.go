package idl

import (
	"fmt"
)

func (p *parser) parseConst() {
	p.advance()

	constType := p.parseType()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected constant name"))
		return
	}

	constName := p.parseIdentifier()

	if p.tok().ID != TokenEquals {
		p.reportError(fmt.Errorf("expected equals"))
		return
	}

	p.advance()

	if p.tok().ID != TokenIdentifier && p.tok().ID != TokenStringLiteral {
		p.reportError(fmt.Errorf("expected constant value"))
		return
	}

	constValue := p.parseValue()

	if p.tok().ID != TokenSemicolon {
		p.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	p.advance()
	p.currentModule.Constants = append(p.currentModule.Constants, Constant{
		Member: Member{
			Name: constName,
			Type: constType,
		},
		Value: constValue,
	})
	if parseDebug {
		fmt.Printf("Got constant: %s of type %s with value %s\n", constName, constType, constValue)
	}
}
