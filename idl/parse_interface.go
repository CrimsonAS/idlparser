package idl

import (
	"fmt"
)

func (pb *parser) parseInterface() {
	pb.advance()

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected interface name"))
		return
	}

	interfaceName := pb.tok().Value
	pb.advance()

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

		inherits := []string{}
		for pb.tok().Id == TokenWord {
			if pb.tok().Id != TokenWord {
				pb.reportError(fmt.Errorf("expected interface inheritance name"))
				return
			}

			inheritsName := pb.tok().Value
			inherits = append(inherits, inheritsName)
			if parseDebug {
				fmt.Printf("Got interface %s inheriting %s\n", interfaceName, inheritsName)
			}
			pb.advance()

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

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected interface member name"))
		return
	}

	memberName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenOpenBracket {
		pb.reportError(fmt.Errorf("expected open bracket"))
		return
	}

	if parseDebug {
		fmt.Printf("Found interface member name %s returning type %s\n", memberName, returnType)
	}
	pb.advance()

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

	param := ""
	for pb.tok().Id == TokenWord || pb.tok().Id == TokenComma {
		switch pb.tok().Id {
		case TokenWord:
			param += pb.tok().Value + " "
		case TokenComma:
			if parseDebug {
				fmt.Printf("Member takes: %s\n", param)
			}
			m.Parameters = append(m.Parameters, param)
			param = ""
		}
		pb.advance()
	}

	if parseDebug {
		fmt.Printf("Member takes: %s\n", param)
	}
	m.Parameters = append(m.Parameters, param)
	pb.currentIface.Methods = append(pb.currentIface.Methods, m)
}
