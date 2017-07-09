package idl

import (
	"fmt"
)

func (pb *ParseBuf) parseTypedef() {
	pb.advance()

	fromName := pb.parseType()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected to name"))
		return
	}

	toName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon, got: %s", pb.tok().id))
		return
	}

	pb.advance()
	fmt.Printf("Typedef: %s to %s\n", fromName, toName)
}
