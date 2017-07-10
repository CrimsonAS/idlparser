package idl

import (
	"fmt"
)

// Handle the start of an enum
// enum MyEnum {
func (pb *ParseBuf) parseEnum() {
	pb.advance()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected enum name"))
		return
	}

	enumName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenOpenBrace {
		pb.reportError(fmt.Errorf("expected enum contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextEnum, enumName)
}

// Handle a member in an enum
// MyValue,
func (pb *ParseBuf) parseEnumMember() {
	// no leading advance, as we start at the name of the enum member.

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected enum value"))
		return
	}

	enumName := pb.tok().value
	pb.advance()

	for pb.tok().id == tokenComma {
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
