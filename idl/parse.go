package idl

import (
	"fmt"
)

const parseDebug = false

type contextId int32

func (ctx contextId) String() string {
	switch ctx {
	case contextGlobal:
		return "global"
	case contextModule:
		return "module"
	case contextStruct:
		return "struct"
	case contextEnum:
		return "enum"
	case contextInterface:
		return "interface"
	}

	return "(wtf)"
}

type context struct {
	id    contextId
	value string
}

const (
	// Outermost
	contextGlobal = iota

	// In a module
	contextModule = iota

	// In a struct
	contextStruct = iota

	// In an enum
	contextEnum = iota

	// In an interface
	contextInterface = iota
)

type ParseBuf struct {
	lb           *LexBuf
	contextStack []context
	ppos         int
	errors       []error
	isEof        bool
}

func (pb *ParseBuf) reportError(err error) {
	// Ignore all errors after EOF, as they are likely bogus (due to our
	// returning a silly token in that case to avoid crashes).
	if !pb.isEof {
		fmt.Printf("Got parse error: %s\n", err)
		pb.errors = append(pb.errors, err)
	}
}

func (pb *ParseBuf) hasError() bool {
	return len(pb.errors) != 0
}

func NewParseBuf(lexBuf *LexBuf) *ParseBuf {
	pb := &ParseBuf{
		lb:    lexBuf,
		isEof: false,
	}
	return pb
}

func (pb *ParseBuf) parseDefineDirective() {
	pb.advance()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	varName := pb.tok().value
	pb.advance()

	if !pb.atEnd() && pb.tok().id == tokenWord {
		varValue := pb.tok().value
		pb.advance()
		fmt.Printf("Define: %s val %s\n", varName, varValue)
	} else {
		fmt.Printf("Define: %s no value\n", varName)
	}
}

func (pb *ParseBuf) parseIncludeDirective() {
	pb.advance()

	if pb.tok().id != tokenStringLiteral {
		pb.reportError(fmt.Errorf("unexpected non-string-literal"))
		return
	}

	fileName := pb.tok().value
	pb.advance()

	fmt.Printf("Included: %s\n", fileName)
}

// The entry point for directives.
func (pb *ParseBuf) parseTokenHash() {
	pb.advance() // skip #

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("unexpected non-word"))
		return
	}

	directive := pb.tok().value

	switch directive {
	case "define":
		pb.parseDefineDirective()
	case "include":
		pb.parseIncludeDirective()
	default:
		pb.reportError(fmt.Errorf("unexpected directive: %s", directive))
	}
}

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

func (pb *ParseBuf) parseType() string {
	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected constant type"))
		return ""
	}

	constType := pb.tok().value
	pb.advance()

	if constType == "unsigned" {
		// consume an additional word
		if pb.tok().id != tokenWord {
			pb.reportError(fmt.Errorf("expected numeric type"))
			return ""
		}

		constType += " " + pb.tok().value
		pb.advance()
	}

	return constType
}

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

func (pb *ParseBuf) parseConst() {
	pb.advance()

	constType := pb.parseType()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected constant name"))
		return
	}

	constName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenEquals {
		pb.reportError(fmt.Errorf("expected equals"))
		return
	}

	pb.advance()

	if pb.tok().id != tokenWord && pb.tok().id != tokenStringLiteral {
		pb.reportError(fmt.Errorf("expected constant value"))
		return
	}

	constValue := ""
	for pb.tok().id == tokenWord || pb.tok().id == tokenStringLiteral {
		constValue += pb.tok().value
		pb.advance()
	}

	if pb.tok().id != tokenSemicolon {
		pb.reportError(fmt.Errorf("expected semicolon"))
		return
	}

	pb.advance()
	fmt.Printf("Got constant: %s of type %s with value %s\n", constName, constType, constValue)
}

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

func (pb *ParseBuf) parseEnumMember() {
	// no leading advance, as we start at the name of the enum member.

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected enum value"))
		return
	}

	enumValue := pb.tok().value
	pb.advance()

	for pb.tok().id == tokenComma {
		// eat the comma(s)
		pb.advance()
	}

	fmt.Printf("Read enum member: %s\n", enumValue)
}

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

func (pb *ParseBuf) parseInterface() {
	pb.advance()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected interface name"))
		return
	}

	interfaceName := pb.tok().value
	pb.advance()

	if pb.tok().id == tokenSemicolon {
		// interface Foo;
		fmt.Printf("Read empty interface %s\n", interfaceName)
		pb.advance()
		return
	}

	if pb.tok().id == tokenOpenBrace {
		// interface Foo {
		fmt.Printf("Read non-inheriting interface %s\n", interfaceName)
		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		return
	}

	if pb.tok().id == tokenColon {
		// interface Foo : Bar {
		pb.advance()

		for pb.tok().id == tokenWord || pb.tok().id == tokenEndLine {
			// This feels a bit nasty...
			for pb.tok().id == tokenEndLine {
				pb.advance()
			}
			if pb.tok().id != tokenWord {
				pb.reportError(fmt.Errorf("expected interface inheritance name"))
				return
			}

			inheritsName := pb.tok().value
			fmt.Printf("Got interface %s inheriting %s\n", interfaceName, inheritsName)
			pb.advance()

			// Multiple inheritance
			if pb.tok().id == tokenComma {
				pb.advance()
			} else if pb.tok().id != tokenOpenBrace {
				pb.reportError(fmt.Errorf("expected open brace"))
				return
			}
		}

		if pb.tok().id != tokenOpenBrace {
			pb.reportError(fmt.Errorf("expected open brace"))
			return
		}

		pb.advance()
		pb.pushContext(contextInterface, interfaceName)
		return
	}

	pb.reportError(fmt.Errorf("invalid interface definition"))
	return
}

func (pb *ParseBuf) parseInterfaceMember() {
	returnType := pb.parseType()

	if pb.tok().id != tokenWord {
		pb.reportError(fmt.Errorf("expected interface member name"))
		return
	}

	memberName := pb.tok().value
	pb.advance()

	if pb.tok().id != tokenOpenBracket {
		pb.reportError(fmt.Errorf("expected open bracket"))
		return
	}

	fmt.Printf("Found interface member name %s returning type %s\n", memberName, returnType)
	pb.advance()

	if pb.tok().id == tokenCloseBracket {
		// void foo();
		pb.advance()
		if pb.tok().id != tokenSemicolon {
			pb.reportError(fmt.Errorf("expected semicolon"))
			return
		}
		pb.advance()
		return
	}

	param := ""
	for pb.tok().id == tokenWord || pb.tok().id == tokenComma || pb.tok().id == tokenEndLine {
		switch pb.tok().id {
		case tokenEndLine:
			// do nothing
		case tokenWord:
			param += pb.tok().value + " "
		case tokenComma:
			fmt.Printf("Member takes: %s\n", param)
			param = ""
		}
		pb.advance()
	}
	fmt.Printf("Member takes: %s\n", param)
}

func (pb *ParseBuf) parseTokenWord() {
	word := pb.tok().value

	switch pb.currentContext().id {
	case contextGlobal:
		fallthrough
	case contextModule:
		switch word {
		case "module":
			pb.parseModule()
		case "typedef":
			pb.parseTypedef()
		case "struct":
			pb.parseStruct()
		case "const":
			pb.parseConst()
		case "enum":
			pb.parseEnum()
		case "interface":
			pb.parseInterface()
		default:
			pb.reportError(fmt.Errorf("unexpected keyword in global/module context: %s", word))
			return
		}
	case contextStruct:
		pb.parseStructMember()
	case contextEnum:
		pb.parseEnumMember()
	case contextInterface:
		pb.parseInterfaceMember()
	}
}

func (pb *ParseBuf) atEnd() bool {
	// ### right?
	return pb.ppos >= len(pb.lb.tokens)-1
}

func (pb *ParseBuf) tok() token {
	if pb.atEnd() {
		pb.isEof = true
		pb.reportError(fmt.Errorf("unexpected EOF"))
		return token{tokenEndLine, ""}
	}
	return pb.lb.tokens[pb.ppos]
}

func (pb *ParseBuf) advance() {
	if parseDebug {
		fmt.Printf("Advancing, ppos was %d, old token %s new token %s\n", pb.ppos, pb.lb.tokens[pb.ppos], pb.lb.tokens[pb.ppos+1])
	}
	pb.ppos += 1
}

func (pb *ParseBuf) Parse() {
	pb.pushContext(contextGlobal, "")

	for !pb.atEnd() && !pb.hasError() {
		tok := pb.tok()
		if parseDebug {
			if len(tok.value) > 0 {
				fmt.Printf("ppos %d Parsing token %s val %s\n", pb.ppos, tok.id, tok.value)
			} else {
				fmt.Printf("ppos %d Parsing token %s\n", pb.ppos, tok.id)
			}
		}

		switch tok.id {
		case tokenHash:
			pb.parseTokenHash()
		case tokenWord:
			pb.parseTokenWord()
		case tokenCloseBrace:
			pb.popContext()
			pb.advance()
		default:
			pb.advance()
		}
	}

	pb.popContext()
	if len(pb.contextStack) > 0 {
		panic("too many contexts")
	}
}

func (pb *ParseBuf) pushContext(ctx contextId, val string) {
	fmt.Printf("Opened context: %s (%s)\n", ctx, val)
	pb.contextStack = append(pb.contextStack, context{ctx, val})
}

func (pb *ParseBuf) popContext() {
	cctx := pb.currentContext()
	fmt.Printf("Closed context: %s (%s)\n", cctx.id, cctx.value)
	pb.contextStack = pb.contextStack[:len(pb.contextStack)-1]
}

func (pb *ParseBuf) currentContext() context {
	return pb.contextStack[len(pb.contextStack)-1]
}
