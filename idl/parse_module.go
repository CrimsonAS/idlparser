package idl

import (
	"fmt"
)

func (pb *parser) parseModule() {
	pb.advance()

	if pb.tok().Id != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected module name"))
		return
	}

	moduleName := pb.parseIdentifier()

	if pb.tok().Id != TokenOpenBrace {
		pb.reportError(fmt.Errorf("expected module contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextModule, moduleName)
}
