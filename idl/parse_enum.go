package idl

import (
	"fmt"
)

// Handle the start of an enum
// enum MyEnum {
func (p *parser) parseEnum() {
	p.advance()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected enum name"))
		return
	}

	enumName := p.parseIdentifier()

	if p.tok().ID != TokenOpenBrace {
		p.reportError(fmt.Errorf("expected enum contents"))
		return
	}

	p.advance()
	p.pushContext(contextEnum, enumName)
}

// Handle a member in an enum
// MyValue,
func (p *parser) parseEnumMember() {
	// no leading advance, as we start at the name of the enum member.

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected enum value"))
		return
	}

	enumName := p.tok().Value
	p.advance()

	for p.tok().ID == TokenComma {
		// eat the comma(s)
		p.advance()
	}

	if parseDebug {
		fmt.Printf("Read enum member: %s\n", enumName)
	}
	p.currentEnum.Members = append(p.currentEnum.Members, Member{
		Name: enumName,
		// ### assign value?
	})
}
