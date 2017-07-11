package idl

import (
	"fmt"
)

// Handle the opening of a struct
// struct Foo {
func (p *parser) parseStruct() {
	p.advance()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected struct name"))
		return
	}

	structName := p.parseIdentifier()

	inherits := []string{}
	switch p.tok().ID {
	case TokenOpenBrace:
		break
	case TokenColon:
		p.advance()
		if p.tok().ID != TokenIdentifier {
			p.reportError(fmt.Errorf("expected struct inheritance"))
			return
		}

		for p.tok().ID == TokenIdentifier {
			name := p.parseIdentifier()
			inherits = append(inherits, name)

			switch p.tok().ID {
			case TokenComma:
				p.advance()
				continue
			case TokenOpenBrace:
				break
			}
		}
	}

	if p.tok().ID != TokenOpenBrace {
		p.reportError(fmt.Errorf("expected struct contents"))
		return
	}

	p.advance()
	p.pushContext(contextStruct, structName)
	p.currentStruct.Inherits = inherits
}

// Handle data members inside a struct
// unsigned long data;
func (p *parser) parseStructMember() {
	typeName := p.parseType()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected member name"))
		return
	}

	memberName := p.parseIdentifier()

	if p.tok().ID != TokenSemicolon {
		p.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	if parseDebug {
		fmt.Printf("Read struct member: %s of type %s\n", memberName, typeName)
	}

	p.currentStruct.Members = append(p.currentStruct.Members, Member{
		Name: memberName,
		Type: typeName,
	})
}
