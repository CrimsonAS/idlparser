package idl

import (
	"fmt"
)

func (p *parser) parseInterface() {
	p.advance()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected interface name"))
		return
	}

	interfaceName := p.parseIdentifier()

	if p.tok().ID == TokenSemicolon {
		// interface Foo;
		if parseDebug {
			fmt.Printf("Read empty interface %s\n", interfaceName)
		}
		p.advance()
		p.pushContext(contextInterface, interfaceName)
		p.popContext() // immediate pop as it's empty, just register in the AST
		return
	}

	if p.tok().ID == TokenOpenBrace {
		// interface Foo {
		if parseDebug {
			fmt.Printf("Read non-inheriting interface %s\n", interfaceName)
		}
		p.advance()
		p.pushContext(contextInterface, interfaceName)
		return
	}

	if p.tok().ID == TokenColon {
		// interface Foo : Bar {
		p.advance()

		if p.tok().ID != TokenIdentifier {
			p.reportError(fmt.Errorf("expected interface inheritance name"))
			return
		}

		inherits := []string{}
		for p.tok().ID == TokenIdentifier {
			inheritsName := p.parseIdentifier()
			inherits = append(inherits, inheritsName)
			if parseDebug {
				fmt.Printf("Got interface %s inheriting %s\n", interfaceName, inheritsName)
			}

			// Multiple inheritance
			if p.tok().ID == TokenComma {
				p.advance()
			} else if p.tok().ID != TokenOpenBrace {
				p.reportError(fmt.Errorf("expected open brace"))
				return
			}
		}

		if p.tok().ID != TokenOpenBrace {
			p.reportError(fmt.Errorf("expected open brace"))
			return
		}

		p.advance()
		p.pushContext(contextInterface, interfaceName)
		p.currentIface.Inherits = inherits
		return
	}

	p.reportError(fmt.Errorf("invalid interface definition"))
	return
}

func (p *parser) parseInterfaceMember() {
	returnType := p.parseType()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected interface member name"))
		return
	}

	memberName := p.parseIdentifier()

	if p.tok().ID != TokenOpenBracket {
		p.reportError(fmt.Errorf("expected open bracket"))
		return
	}
	p.advance()

	if parseDebug {
		fmt.Printf("Found interface member name %s returning type %s\n", memberName, returnType)
	}

	m := Method{
		Name:        memberName,
		ReturnValue: returnType,
	}

	if p.tok().ID == TokenCloseBracket {
		// void foo();
		p.advance()
		if p.tok().ID != TokenSemicolon {
			p.reportError(fmt.Errorf("expected semicolon"))
			return
		}
		p.advance()
		p.currentIface.Methods = append(p.currentIface.Methods, m)
		return
	}

	for {
		if p.tok().ID != TokenIdentifier {
			p.reportError(fmt.Errorf("expected direction"))
			return
		}

		switch p.tok().Value {
		case keywordIn:
		case keywordOut:
		case keywordInOut:
			break
		default:
			p.reportError(fmt.Errorf("unexpected direction"))
			return
		}

		direction := p.tok().Value
		p.advance()

		typeName := p.parseType()

		paramName := ""

		// Allow: "in foo bar" and "in foo"
		if p.tok().ID == TokenIdentifier {
			paramName = p.parseIdentifier()
		}

		full := fmt.Sprintf("%s %s %s", direction, typeName, paramName)

		if parseDebug {
			fmt.Printf("Member takes: %s\n", full)
		}
		m.Parameters = append(m.Parameters, full)

		switch p.tok().ID {
		case TokenCloseBracket:
			goto out
		case TokenComma:
			p.advance()
			continue
		}
	}

out:
	p.currentIface.Methods = append(p.currentIface.Methods, m)
}
