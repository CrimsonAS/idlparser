package idl

import (
	"fmt"
)

// Handle the opening of a struct
// struct Foo {
func (pb *ParseBuf) parseStruct() {
	pb.advance()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected struct name"))
		return
	}

	structName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenOpenBrace {
		pb.reportError(fmt.Errorf("expected struct contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextStruct, structName)
}

// Handle data members inside a struct
// unsigned long data;
func (pb *ParseBuf) parseStructMember() {
	typeName := pb.parseType()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected member name"))
		return
	}

	memberName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	pb.advance()
	fmt.Printf("Read struct member: %s of type %s\n", memberName, typeName)
}
