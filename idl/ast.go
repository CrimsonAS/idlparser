package idl

// Member provides a generic representation of a member in the AST
type Member struct {
	// The name of the member
	Name string

	// The type of the member (e.g. "unsigned long")
	Type string
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

// Method represents the contents of a method in an Interface in the AST
// For instance, "void foo(inout int foo)"
type Method struct {
	// The name of the method (e.g. foo)
	Name string

	// The return value of the method (e.g. void)
	ReturnValue string

	// The parameters of the method (e.g. "inout int foo")
	Parameters []string
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
