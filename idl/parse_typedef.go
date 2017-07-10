package idl

import (
	"fmt"
)

func (pb *parser) parseTypedef() {
	pb.advance()

	fromName := pb.parseType()

	if pb.tok().Id != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected to name"))
		return
	}

	toName := pb.parseIdentifier()

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
