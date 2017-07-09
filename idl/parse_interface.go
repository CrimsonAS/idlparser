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
		fmt.Printf("Read empty interface %s\n", interfaceName)
		pb.advance()
		return
	}

	if pb.tok().id == tokenOpenBrace {
		// interface Foo {
		fmt.Printf("Read non-inheriting interface %s\n", interfaceName)
		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		return
	}

	if pb.tok().id == tokenColon {
		// interface Foo : Bar {
		pb.advance()

		for pb.tok().id == tokenWord || pb.tok().id == tokenEndLine {
			// This feels a bit nasty...
			for pb.tok().id == tokenEndLine {
				pb.advance()
			}
			if pb.tok().id != tokenWord {
				pb.reportError(fmt.Errorf("expected interface inheritance name"))
				return
			}

			inheritsName := pb.tok().value
			fmt.Printf("Got interface %s inheriting %s\n", interfaceName, inheritsName)
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

	fmt.Printf("Found interface member name %s returning type %s\n", memberName, returnType)
	pb.advance()

	if pb.tok().id == tokenCloseBracket {
		// void foo();
		pb.advance()
		if pb.tok().id != tokenSemicolon {
			pb.reportError(fmt.Errorf("expected semicolon"))
			return
		}
		pb.advance()
		return
	}

	param := ""
	for pb.tok().id == tokenWord || pb.tok().id == tokenComma || pb.tok().id == tokenEndLine {
		switch pb.tok().id {
		case tokenEndLine:
			// do nothing
		case tokenWord:
			param += pb.tok().value + " "
		case tokenComma:
			fmt.Printf("Member takes: %s\n", param)
			param = ""
		}
		pb.advance()
	}
	fmt.Printf("Member takes: %s\n", param)
}
