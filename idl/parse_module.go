package idl

import (
	"fmt"
)

func (pb *ParseBuf) parseModule() {
	pb.advance()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected module name"))
		return
	}

	moduleName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenOpenBrace {
		pb.reportError(fmt.Errorf("expected module contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextModule, moduleName)
}
