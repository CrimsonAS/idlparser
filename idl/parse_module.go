package idl

import (
	"fmt"
)

func (pb *Parser) parseModule() {
	pb.advance()

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected module name"))
		return
	}

	moduleName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenOpenBrace {
		pb.reportError(fmt.Errorf("expected module contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextModule, moduleName)
}
