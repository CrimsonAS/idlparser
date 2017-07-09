package main

import (
	"./idl"
	"fmt"
	"io/ioutil"
)

func checkErr(err error, what string) {
	if err != nil {
		fmt.Errorf("error! %s (%s)", err, what)
	}
}

func main() {
	b, err := ioutil.ReadFile("dds_dcps.idl")
	checkErr(err, "reading file")
	lb := idl.NewLexBuf(b)
	lb.Lex()
	pb := idl.NewParseBuf(lb)
	pb.Parse()
}
