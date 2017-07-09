package idl

import (
	"fmt"
)

func (pb *ParseBuf) parseConst() {
	pb.advance()

	constType := pb.parseType()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected constant name"))
		return
	}

	constName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenEquals {
		pb.reportError(fmt.Errorf("expected equals"))
		return
	}

	pb.advance()

	if pb.tok().id != tokenWord && pb.tok().id != tokenStringLiteral {
		pb.reportError(fmt.Errorf("expected constant value"))
		return
	}

	constValue := ""
	for pb.tok().id == tokenWord || pb.tok().id == tokenStringLiteral {
		constValue += pb.tok().value
		pb.advance()
	}

	if pb.tok().id != tokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	pb.advance()
	fmt.Printf("Got constant: %s of type %s with value %s\n", constName, constType, constValue)
}
