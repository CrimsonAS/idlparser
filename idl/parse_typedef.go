package idl

import (
	"fmt"
)

func (pb *Parser) parseTypedef() {
	pb.advance()

	fromName := pb.parseType()

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected to name"))
		return
	}

	toName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon, got: %s", pb.tok().Id))
		return
	}

	pb.advance()
	pb.currentModule.TypeDefs = append(pb.currentModule.TypeDefs, TypeDef{
		Name: toName,
		Type: fromName,
	})
	if parseDebug {
		fmt.Printf("Typedef: %s to %s\n", fromName, toName)
	}
}
