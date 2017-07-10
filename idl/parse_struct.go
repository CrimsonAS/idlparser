package idl

import (
	"fmt"
)

// Handle the opening of a struct
// struct Foo {
func (pb *ParseBuf) parseStruct() {
	pb.advance()

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected struct name"))
		return
	}

	structName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenOpenBrace {
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

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected member name"))
		return
	}

	memberName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	pb.advance()

	if parseDebug {
		fmt.Printf("Read struct member: %s of type %s\n", memberName, typeName)
	}

	pb.currentStruct.Members = append(pb.currentStruct.Members, Member{
		Name: memberName,
		Type: typeName,
	})
}
