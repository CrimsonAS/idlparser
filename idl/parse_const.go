package idl

import (
	"fmt"
)

func (pb *Parser) parseConst() {
	pb.advance()

	constType := pb.parseType()

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected constant name"))
		return
	}

	constName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenEquals {
		pb.reportError(fmt.Errorf("expected equals"))
		return
	}

	pb.advance()

	if pb.tok().Id != TokenWord && pb.tok().Id != TokenStringLiteral {
		pb.reportError(fmt.Errorf("expected constant value"))
		return
	}

	constValue := ""
	for pb.tok().Id == TokenWord || pb.tok().Id == TokenStringLiteral {
		constValue += pb.tok().Value
		pb.advance()
	}

	if pb.tok().Id != TokenSemicolon {
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
