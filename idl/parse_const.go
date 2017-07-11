package idl

import (
	"fmt"
)

func (pb *parser) parseConst() {
	pb.advance()

	constType := pb.parseType()

	if pb.tok().ID != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected constant name"))
		return
	}

	constName := pb.parseIdentifier()

	if pb.tok().ID != TokenEquals {
		pb.reportError(fmt.Errorf("expected equals"))
		return
	}

	pb.advance()

	if pb.tok().ID != TokenIdentifier && pb.tok().ID != TokenStringLiteral {
		pb.reportError(fmt.Errorf("expected constant value"))
		return
	}

	constValue := pb.parseValue()

	if pb.tok().ID != TokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	pb.advance()
	pb.currentModule.Constants = append(pb.currentModule.Constants, Constant{
		Name: constName,
		Type: constType,
	})
	if parseDebug {
		fmt.Printf("Got constant: %s of type %s with value %s\n", constName, constType, constValue)
	}
}
