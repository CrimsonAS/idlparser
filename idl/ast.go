package idl

import (
	"fmt"
	"strings"
)

// Type provides a parsed representation of an IDL type.
type Type struct {
	// The name of the type, e.g. "boolean", or "sequence" in "sequence<string>"
	Name string

	// The number of this type, used for fixed size arrays, e.g: int foo[3] --
	// Quantity will be 3. If there is no quantity, this will be nil.
	Quantity *int

	// Any parameters of the type if the type is a templated one (e.g. "string"
	// in "sequence<string>")
	TemplateParameters []Type
}

func (t Type) String() string {
	tparams := []string{}
	for _, tp := range t.TemplateParameters {
		tparams = append(tparams, tp.Name)
	}
	if len(tparams) > 0 {
		return fmt.Sprintf("%s<%s>", t.Name, strings.Join(tparams, ", "))
	}
	return fmt.Sprintf("%s", t.Name)
}

// Member provides a generic representation of a member in the AST
type Member struct {
	// The name of the member
	Name string

	// The type of the member (e.g. "unsigned long")
	Type Type
}

// TypeDef represents a typedef in the AST
type TypeDef Member

// Constant represents a constant in the AST
type Constant struct {
	// The meta-information about the constant
	Member

	// The value of the constant
	Value string
}

// Struct represents a struct in the AST
type Struct struct {
	// The name of the struct
	Name string

	// What struct this struct inherits
	Inherits []string

	// The members inside this struct
	Members []Member
}

// Enum represents an enum in the AST
type Enum struct {
	// The name of the enum
	Name string

	// The members inside this enum
	Members []Member
}

// MethodParameter is a specialization of Type to provide the direction of the
// type.
type MethodParameter struct {
	Type

	// in, out, inout
	Direction string
}

func (t MethodParameter) String() string {
	return fmt.Sprintf("%s %s", t.Direction, t.Type)
}

// Method represents the contents of a method in an Interface in the AST
// For instance, "void foo(inout int foo)"
type Method struct {
	// The name of the method (e.g. foo)
	Name string

	// The return value of the method (e.g. void)
	ReturnValue Type

	// The parameters of the method (e.g. "inout int foo")
	Parameters []MethodParameter
}

// Interface represents an interface in the AST
type Interface struct {
	// The name of the interface
	Name string

	// What interfaces this interface inherits
	Inherits []string

	// What methods this interface provides
	Methods []Method
}

// Module is the base type of the AST generated from the parsed IDL.
// It contains everything in the file in an easily accessible form.
//
// It can be nested (as in module Outer { module Inner { } }). An
// empty "root" module will contain everything not inside a module
// declaration.
type Module struct {
	// The name of the module
	Name string

	// The parent module
	parent *Module

	// Modules inside this module
	Modules []Module

	// Interfaces inside this module
	Interfaces []Interface

	// All typedefs in this module
	TypeDefs []TypeDef

	// All structs in this module
	Structs []Struct

	// All constants in this module
	Constants []Constant

	// All enums in this module
	Enums []Enum
}
