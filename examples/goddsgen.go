package main

import (
	"../idl"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func checkErr(err error, what string) {
	if err != nil {
		fmt.Errorf("error! %s (%s)", err, what)
		os.Exit(-1)
	}
}

func main() {
	file := flag.String("file", "dds_dcps.idl", "file to parse")
	baseModule := flag.String("module", "Dds", "base module to generate")
	flag.Parse()

	if file == nil {
		fmt.Printf("Need a filename\n")
		os.Exit(-1)
	}

	b, err := ioutil.ReadFile(*file)
	checkErr(err, "reading file")
	tokens, err := idl.Lex(b)
	checkErr(err, "lexing")
	module, err := idl.Parse(tokens)
	checkErr(err, "parsing")
	module.Name = *baseModule
	generateModule(module)
}

// Turn an IDL type (like "sequence<Foo>") into a Go type ("[]Foo")
func idlTypeToGoType(idlType idl.Type) string {
	rtype := ""

	if idlType.Quantity != nil {
		rtype = fmt.Sprintf("[%d]", *idlType.Quantity)
	}

	if idlType.Name == "unsigned long" {
		rtype += "uint32"
	} else if idlType.Name == "long" {
		rtype += "int32"
	} else if idlType.Name == "sequence" {
		if len(idlType.TemplateParameters) == 1 {
			rtype += fmt.Sprintf("[]%s", idlType.TemplateParameters[0].Name)
		} else if len(idlType.TemplateParameters) == 2 {
			rtype += fmt.Sprintf("[%s]%s", idlType.TemplateParameters[2].Name, idlType.TemplateParameters[0].Name)
		} else {
			panic("too many params")
		}
	} else {
		rtype += idlType.Name
	}

	return rtype
}

// Turn an IDL var name (like "foo_bar") into a Go-friendly CamelCase one (FooBar)
func identifierToGoIdentifier(identifier string) string {
	nid := strings.ToUpper(string(identifier[0])) + identifier[1:]
	for idx := strings.Index(nid, "_"); idx > 0; idx = strings.Index(nid, "_") {
		nid = nid[:idx] + strings.ToUpper(nid[idx+1:idx+2]) + nid[idx+2:]
	}
	return nid
}

// ### todo: write this to disk, not stdout. nest the generated code in
// directories, so:
//
// dds_generated.go
//     sub_module/dds_generated.go
//
//... etc. One Go module per IDL module.
func generateModule(m idl.Module) {
	fmt.Printf("package %s\n", m.Name)
	fmt.Printf("\n\n")
	for _, t := range m.Constants {
		fmt.Printf("const %s = %s\n", t.Name, t.Value)
	}

	fmt.Printf("\n\n")

	for _, t := range m.TypeDefs {
		fmt.Printf("type %s %s\n", t.Name, idlTypeToGoType(t.Type))
	}
	fmt.Printf("\n\n")

	for _, t := range m.Structs {
		fmt.Printf("type %s struct {\n", t.Name)

		for _, t2 := range t.Members {
			fmt.Printf("\t%s %s\n", identifierToGoIdentifier(t2.Name), idlTypeToGoType(t2.Type))

		}

		fmt.Printf("}\n")
	}
	fmt.Printf("\n\n")

	for _, t := range m.Modules {
		generateModule(t)
	}
}
