package idl

// A generic representation of a member
type Member struct {
	// The name of the member
	Name string

	// The type of the member (e.g. "unsigned long")
	Type string
}

type TypeDef Member
type Constant Member

// The contents of a struct{}
type Struct struct {
	// The name of the struct
	Name string

	// The members inside this struct
	Members []Member
}

// The contents of an enum{}
type Enum struct {
	// The name of the enum
	Name string

	// The members inside this enum
	Members []Member
}

// The contents of a method in an Interface
// For instance, "void foo(inout int foo)"
type Method struct {
	// The name of the method (e.g. foo)
	Name string

	// The return value of the method (e.g. void)
	ReturnValue string

	// The parameters of the method (e.g. "inout int foo")
	Parameters []string
}

// The contents of an interface {}
type Interface struct {
	// The name of the interface
	Name string

	// What interfaces this interface inherits
	Inherits []string

	// What methods this interface provides
	Methods []Method
}

// The root of all the IDL
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
