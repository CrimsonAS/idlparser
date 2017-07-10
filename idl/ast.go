package idl

// A generic representation of a member
type Member struct {
	// The name of the member
	Name string

	// The type of the member (e.g. "unsigned long")
	Type string
}

// Represents a typedef.
type TypeDef Member

// Represents a constant.
type Constant Member

// Represents a struct.
type Struct struct {
	// The name of the struct
	Name string

	// The members inside this struct
	Members []Member
}

// Represents an enum.
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

// Represents an interface.
type Interface struct {
	// The name of the interface
	Name string

	// What interfaces this interface inherits
	Inherits []string

	// What methods this interface provides
	Methods []Method
}

// A module is the base type of the contents of the IDL. It contains everything
// in the file in an easily accessible form.
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
