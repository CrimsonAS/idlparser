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
	printModule(module)
}

var recurse = 0

func printModule(m idl.Module) {
	tabs := ""
	for i := 0; i < recurse; i++ {
		tabs += "\t"
	}

	recurse++
	defer func() { recurse-- }()

	fmt.Printf("%sModule: %s\n", tabs, m.Name)
	fmt.Printf("%s\t%s Interfaces:\n", tabs, m.Name)
	for _, t := range m.Interfaces {
		fmt.Printf("%s\t\t%s (: %s)\n", tabs, t.Name, strings.Join(t.Inherits, ", "))
		for _, t2 := range t.Methods {
			fmt.Printf("%s\t\t\t%s %s(%s)\n", tabs, t2.ReturnValue, t2.Name, t2.Parameters)
		}
	}
	fmt.Printf("%s\t%s Constants:\n", tabs, m.Name)
	for _, t := range m.Constants {
		fmt.Printf("%s\t\t%s (%s)\n", tabs, t.Name, t.Type)
	}
	fmt.Printf("%s\t%s TypeDefs:\n", tabs, m.Name)
	for _, t := range m.TypeDefs {
		fmt.Printf("%s\t\t%s (%s)\n", tabs, t.Name, t.Type)
	}
	fmt.Printf("%s\t%s Enums:\n", tabs, m.Name)
	for _, t := range m.Enums {
		fmt.Printf("%s\t\t%s\n", tabs, t.Name)
		for _, t2 := range t.Members {
			fmt.Printf("%s\t\t\t%s (%s)\n", tabs, t2.Name, t2.Type)
		}
	}
	fmt.Printf("%s\t%s Unions:\n", tabs, m.Name)
	for _, t := range m.Unions {
		fmt.Printf("%s\t\t%s (on type %s)\n", tabs, t.Name, t.Discriminant)
		for _, t2 := range t.Members {
			fmt.Printf("%s\t\t\tcase %s (%s %s)\n", tabs, t2.CaseValue, t2.MemberType, t2.MemberName)
		}
	}
	fmt.Printf("%s\t%s Structs:\n", tabs, m.Name)
	for _, t := range m.Structs {
		fmt.Printf("%s\t\t%s\n", tabs, t.Name)
		for _, t2 := range t.Members {
			fmt.Printf("%s\t\t\t%s (%s)\n", tabs, t2.Name, t2.Type)
		}
	}

	for _, t := range m.Modules {
		printModule(t)
	}
}
