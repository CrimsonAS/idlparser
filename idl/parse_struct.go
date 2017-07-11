package idl

import (
	"fmt"
)

// Handle the opening of a struct
// struct Foo {
func (pb *parser) parseStruct() {
	pb.advance()

	if pb.tok().ID != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected struct name"))
		return
	}

	structName := pb.parseIdentifier()

	inherits := []string{}
	switch pb.tok().ID {
	case TokenOpenBrace:
		break
	case TokenColon:
		pb.advance()
		if pb.tok().ID != TokenIdentifier {
			pb.reportError(fmt.Errorf("expected struct inheritance"))
			return
		}

		for pb.tok().ID == TokenIdentifier {
			name := pb.parseIdentifier()
			inherits = append(inherits, name)

			switch pb.tok().ID {
			case TokenComma:
				pb.advance()
				continue
			case TokenOpenBrace:
				break
			}
		}
	}

	if pb.tok().ID != TokenOpenBrace {
		pb.reportError(fmt.Errorf("expected struct contents"))
		return
	}

	pb.advance()
	pb.pushContext(contextStruct, structName)
	pb.currentStruct.Inherits = inherits
}

// Handle data members inside a struct
// unsigned long data;
func (pb *parser) parseStructMember() {
	typeName := pb.parseType()

	if pb.tok().ID != TokenIdentifier {
		pb.reportError(fmt.Errorf("expected member name"))
		return
	}

	memberName := pb.parseIdentifier()

	if pb.tok().ID != TokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	if parseDebug {
		fmt.Printf("Read struct member: %s of type %s\n", memberName, typeName)
	}

	pb.currentStruct.Members = append(pb.currentStruct.Members, Member{
		Name: memberName,
		Type: typeName,
	})
}
