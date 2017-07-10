package idl

import (
	"fmt"
)

// Handle the start of an enum
// enum MyEnum {
func (pb *parser) parseEnum() {
	pb.advance()

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected enum name"))
		return
	}

	enumName := pb.tok().Value
	pb.advance()

	if pb.tok().Id != TokenOpenBrace {
		pb.reportError(fmt.Errorf("expected enum contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextEnum, enumName)
}

// Handle a member in an enum
// MyValue,
func (pb *parser) parseEnumMember() {
	// no leading advance, as we start at the name of the enum member.

	if pb.tok().Id != TokenWord {
		pb.reportError(fmt.Errorf("expected enum value"))
		return
	}

	enumName := pb.tok().Value
	pb.advance()

	for pb.tok().Id == TokenComma {
		// eat the comma(s)
		pb.advance()
	}

	if parseDebug {
		fmt.Printf("Read enum member: %s\n", enumName)
	}
	pb.currentEnum.Members = append(pb.currentEnum.Members, Member{
		Name: enumName,
		// ### assign value?
	})
}
