package idl

import (
	"fmt"
)

func (pb *parser) parseInterface() {
	pb.advance()

	if pb.tok().Id != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected interface name"))
		return
	}

	interfaceName := pb.parseIdentifier()

	if pb.tok().Id == TokenSemicolon {
		// interface Foo;
		if parseDebug {
			fmt.Printf("Read empty interface %s\n", interfaceName)
		}
		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		pb.popContext() // immediate pop as it's empty, just register in the AST
		return
	}

	if pb.tok().Id == TokenOpenBrace {
		// interface Foo {
		if parseDebug {
			fmt.Printf("Read non-inheriting interface %s\n", interfaceName)
		}
		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		return
	}

	if pb.tok().Id == TokenColon {
		// interface Foo : Bar {
		pb.advance()

		if pb.tok().Id != TokenIdentifier {
			pb.reportError(fmt.Errorf("expected interface inheritance name"))
			return
		}

		inherits := []string{}
		for pb.tok().Id == TokenIdentifier {
			inheritsName := pb.parseIdentifier()
			inherits = append(inherits, inheritsName)
			if parseDebug {
				fmt.Printf("Got interface %s inheriting %s\n", interfaceName, inheritsName)
			}

			// Multiple inheritance
			if pb.tok().Id == TokenComma {
				pb.advance()
			} else if pb.tok().Id != TokenOpenBrace {
				pb.reportError(fmt.Errorf("expected open brace"))
				return
			}
		}

		if pb.tok().Id != TokenOpenBrace {
			pb.reportError(fmt.Errorf("expected open brace"))
			return
		}

		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		pb.currentIface.Inherits = inherits
		return
	}

	pb.reportError(fmt.Errorf("invalid interface definition"))
	return
}

func (pb *parser) parseInterfaceMember() {
	returnType := pb.parseType()

	if pb.tok().Id != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected interface member name"))
		return
	}

	memberName := pb.parseIdentifier()

	if pb.tok().Id != TokenOpenBracket {
		pb.reportError(fmt.Errorf("expected open bracket"))
		return
	}
	pb.advance()

	if parseDebug {
		fmt.Printf("Found interface member name %s returning type %s\n", memberName, returnType)
	}

	m := Method{
		Name:        memberName,
		ReturnValue: returnType,
	}

	if pb.tok().Id == TokenCloseBracket {
		// void foo();
		pb.advance()
		if pb.tok().Id != TokenSemicolon {
			pb.reportError(fmt.Errorf("expected semicolon"))
			return
		}
		pb.advance()
		pb.currentIface.Methods = append(pb.currentIface.Methods, m)
		return
	}

	for {
		if pb.tok().Id != TokenKeyword {
			pb.reportError(fmt.Errorf("expected direction"))
			return
		}

		switch pb.tok().Value {
		case keywordIn:
		case keywordOut:
		case keywordInOut:
			break
		default:
			pb.reportError(fmt.Errorf("unexpected direction"))
			return
		}

		direction := pb.tok().Value
		pb.advance()

		typeName := pb.parseType()

		paramName := ""

		// Allow: "in foo bar" and "in foo"
		if pb.tok().Id == TokenIdentifier {
			paramName = pb.parseIdentifier()
		}

		full := fmt.Sprintf("%s %s %s", direction, typeName, paramName)

		if parseDebug {
			fmt.Printf("Member takes: %s\n", full)
		}
		m.Parameters = append(m.Parameters, full)

		switch pb.tok().Id {
		case TokenCloseBracket:
			goto out
		case TokenComma:
			pb.advance()
			continue
		}
	}

out:
	pb.currentIface.Methods = append(pb.currentIface.Methods, m)
}
