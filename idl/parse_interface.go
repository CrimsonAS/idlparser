package idl

import (
	"fmt"
)

func (pb *ParseBuf) parseInterface() {
	pb.advance()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected interface name"))
		return
	}

	interfaceName := pb.tok().value
	pb.advance()

	if pb.tok().id == tokenSemicolon {
		// interface Foo;
		if parseDebug {
			fmt.Printf("Read empty interface %s\n", interfaceName)
		}
		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		pb.popContext() // immediate pop as it's empty, just register in the AST
		return
	}

	if pb.tok().id == tokenOpenBrace {
		// interface Foo {
		if parseDebug {
			fmt.Printf("Read non-inheriting interface %s\n", interfaceName)
		}
		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		return
	}

	if pb.tok().id == tokenColon {
		// interface Foo : Bar {
		pb.advance()

		inherits := []string{}
		for pb.tok().id == tokenWord {
			if pb.tok().id != tokenWord {
				pb.reportError(fmt.Errorf("expected interface inheritance name"))
				return
			}

			inheritsName := pb.tok().value
			inherits = append(inherits, inheritsName)
			if parseDebug {
				fmt.Printf("Got interface %s inheriting %s\n", interfaceName, inheritsName)
			}
			pb.advance()

			// Multiple inheritance
			if pb.tok().id == tokenComma {
				pb.advance()
			} else if pb.tok().id != tokenOpenBrace {
				pb.reportError(fmt.Errorf("expected open brace"))
				return
			}
		}

		if pb.tok().id != tokenOpenBrace {
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

func (pb *ParseBuf) parseInterfaceMember() {
	returnType := pb.parseType()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected interface member name"))
		return
	}

	memberName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenOpenBracket {
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

	if pb.tok().id == tokenCloseBracket {
		// void foo();
		pb.advance()
		if pb.tok().id != tokenSemicolon {
			pb.reportError(fmt.Errorf("expected semicolon"))
			return
		}
		pb.advance()
		pb.currentIface.Methods = append(pb.currentIface.Methods, m)
		return
	}

	param := ""
	for pb.tok().id == tokenWord || pb.tok().id == tokenComma {
		switch pb.tok().id {
		case tokenWord:
			param += pb.tok().value + " "
		case tokenComma:
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
