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

	n := idlType.Name
	if idx := strings.Index(n, "::"); idx >= 0 {
		// strip off namespace prefix
		n = n[idx+2:]
	}

	if n == "unsigned long" {
		rtype += "uint32"
	} else if n == "boolean" {
		rtype += "bool"
	} else if n == "long long" {
		rtype += "int64"
	} else if n == "long" {
		rtype += "int32"
	} else if n == "float" {
		rtype += "float32"
	} else if n == "double" {
		rtype += "float64"
	} else if n == "sequence" {
		nestedType := idl.Type{Name: idlType.TemplateParameters[0].Name}
		if len(idlType.TemplateParameters) == 1 {
			rtype += fmt.Sprintf("[]%s", idlTypeToGoType(nestedType))
		} else if len(idlType.TemplateParameters) == 2 {
			rtype += fmt.Sprintf("[%s]%s", idlType.TemplateParameters[1].Name, idlTypeToGoType(nestedType))
		} else {
			panic("too many params")
		}
	} else {
		rtype += n
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
	if m.Parent == nil {
		fmt.Printf("package main\n")
		//fmt.Printf("package %s\n", m.Name)
	}

	fmt.Printf("\n\n")
	for _, t := range m.Constants {
		fmt.Printf("const %s = %s\n", t.Name, t.Value)
	}

	fmt.Printf("\n\n")

	fmt.Printf("// TypeDefs\n")
	for _, t := range m.TypeDefs {
		fmt.Printf("type %s %s\n", t.Name, idlTypeToGoType(t.Type))
	}
	fmt.Printf("\n\n")

	// ### this needs a lot of fleshing out i'm sure
	fmt.Printf("// Unions\n")
	for _, t := range m.Unions {
		fmt.Printf("type %s struct {\n", t.Name)
		fmt.Printf("}\n")

		for _, t2 := range t.Members {
			fmt.Printf("func (u *%s) %s() %s {", t.Name, identifierToGoIdentifier(t2.MemberName), idlTypeToGoType(t2.MemberType))
			fmt.Printf("return %s{}", idlTypeToGoType(t2.MemberType))
			fmt.Printf("}\n")
		}
	}
	fmt.Printf("\n\n")

	fmt.Printf("// Enums\n")
	for _, t := range m.Enums {
		fmt.Printf("type %s int32\n", t.Name)
		fmt.Printf("const (\n")

		for idx, t2 := range t.Members {
			if idx == 0 {
				fmt.Printf("\t%s%s = iota\n", t.Name, t2.Name)
			} else {
				fmt.Printf("\t%s%s\n", t.Name, t2.Name)
			}
		}

		fmt.Printf(")\n")
	}
	fmt.Printf("\n\n")

	fmt.Printf("// Structs\n")
	for _, t := range m.Structs {
		fmt.Printf("type %s struct {\n", t.Name)
		for _, t2 := range t.Inherits {
			fmt.Printf("\t%s\n", t2)

		}

		for _, t2 := range t.Members {
			fmt.Printf("\t%s %s\n", identifierToGoIdentifier(t2.Name), idlTypeToGoType(t2.Type))

		}

		fmt.Printf("}\n")
		fmt.Printf("type %sSeq []%s\n", t.Name, t.Name)

		fmt.Printf("type %sTypeSupport struct {\n", t.Name)
		fmt.Printf("}\n")

		fmt.Printf("func (s *%sTypeSupport) RegisterType(participant dds.DomainParticipant, type_name string) dds.ReturnCode_t {\n", t.Name)
		fmt.Printf("\tparticipant.RegisterType(%s{}, \"%s\")\n", t.Name, t.Name)
		fmt.Printf("}\n")
		fmt.Printf("func (s *%sTypeSupport) GetTypeName() string {\n", t.Name)
		fmt.Printf("\treturn \"%s\"\n", t.Name)
		fmt.Printf("}\n")

		fmt.Printf("type %sDataWriter struct {\n", t.Name)
		fmt.Printf("\tw dds.DataWriter\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) RegisterInstance(instance_data %s) dds.InstanceHandle_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.RegisterInstance(instance_data)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) RegisterInstanceWithTimestamp(instance_data %s, source_timestamp dds.Time_t) dds.InstanceHandle_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.RegisterInstanceWithTimestamp(instance_data, source_timestamp)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) UnregisterInstance(instance_data %s, handle dds.InstanceHandle_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.UnregisterInstance(instance_data)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) UnregisterInstanceWithTimestamp(instance_data %s, handle dds.InstanceHandle_t, source_timestamp dds.Time_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.UnregisterInstanceWithTimestamp(instance_data, handle, source_timestamp)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dw *%sDataWriter) Write(instance_data %s, handle dds.InstanceHandle_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.Write(instance_data, handle)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) WriteWithTimestamp(instance_data %s, handle dds.InstanceHandle_t, source_timestamp dds.Time_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.WriteWithTimestamp(instance_data, handle, source_timestamp)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dw *%sDataWriter) Dispose(instance_data %s, instance_handle dds.InstanceHandle_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.Dispose(instance_data, instance_handle)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) DisposeWithTimestamp(instance_data %s, instance_handle dds.InstanceHandle_t, source_timestamp dds.Time_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.DisposeWithTimestamp(instance_data, instance_handle, source_timestamp)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dw *%sDataWriter) GetKeyValue(instance_data *%s, handle dds.InstanceHandle_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.GetKeyValue(instance_data, handle)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dw *%sDataWriter) LookupInstance(key_holder %s) dds.InstanceHandle_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dw.w.LookupInstance(key_holder)\n")
		fmt.Printf("}\n")

		fmt.Printf("type %sDataReader struct {\n", t.Name)
		fmt.Printf("\tr dds.DataReader\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) Read(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, sample_states dds.SampleStateMask, view_states dds.ViewStateMask, instance_states dds.InstanceStateMask) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.Read(data_values, sample_infos, max_samples, sample_states, view_states, instance_states)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) Take(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, sample_states dds.SampleStateMask, view_states dds.ViewStateMask, instance_states dds.InstanceStateMask) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.Take(data_values, sample_infos, max_samples, sample_states, view_states, instance_states)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) ReadWithCondition(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, a_condition dds.ReadCondition) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.ReadWithCondition(data_values, sample_infos, max_samples, a_condition)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) TakeWithCondition(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, a_condition dds.ReadCondition) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.TakeWithCondition(data_values, sample_infos, max_samples, a_condition)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) ReadNextSample(data_values *[]%s, sample_infos *[]dds.SampleInfo) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.ReadNextSample(data_values, sample_infos)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) TakeNextSample(data_values *[]%s, sample_infos *[]dds.SampleInfo) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.TakeNextSample(data_values, sample_infos)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) ReadInstance(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, a_handle dds.InstanceHandle_t, sample_states dds.SampleStateMask, view_states dds.ViewStateMask, instance_states dds.InstanceStateMask) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.ReadInstance(data_values, sample_infos, max_samples, a_handle, sample_states, view_states, instance_states)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) TakeInstance(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, a_handle dds.InstanceHandle_t, sample_states dds.SampleStateMask, view_states dds.ViewStateMask, instance_states dds.InstanceStateMask) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.TakeInstance(data_values, sample_infos, max_samples, a_handle, sample_states, view_states, instance_states)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) ReadNextInstance(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, a_handle dds.InstanceHandle_t, sample_states dds.SampleStateMask, view_states dds.ViewStateMask, instance_states dds.InstanceStateMask) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.ReadNextInstance(data_values, sample_infos, max_samples, a_handle, sample_states, view_states, instance_states)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) TakeNextInstance(data_values *[]%s, sample_info *[]dds.SampleInfo, max_samples int32, a_handle dds.InstanceHandle_t, sample_states dds.SampleStateMask, view_states dds.ViewStateMask, instance_states dds.InstanceStateMask) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.TakeNextInstance(data_values, sample_infos, max_samples, a_handle, sample_states, view_states, instance_states)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) ReadNextInstanceWithCondition(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, previous_handle dds.InstanceHandle_t, a_condition dds.ReadCondition) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.ReadNextInstanceWithCondition(data_values, sample_infos, max_samples, previous_handle, a_condition)\n")
		fmt.Printf("}\n")
		fmt.Printf("func (dr *%sDataReader) TakeNextInstanceWithCondition(data_values *[]%s, sample_infos *[]dds.SampleInfo, max_samples int32, previous_handle dds.InstanceHandle_t, a_condition dds.ReadCondition) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.TakeNextInstanceWithCondition(data_values, sample_infos, max_samples, previous_handle, a_condition)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) ReturnLoan(data_values *[]%s, sample_infos *[]dds.SampleInfo) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.ReturnLoan(data_values, sample_infos)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) GetKeyValue(key_holder *%s, handle dds.InstanceHandle_t) dds.ReturnCode_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.GetKeyValue(key_holder, handle)\n")
		fmt.Printf("}\n")

		fmt.Printf("func (dr *%sDataReader) LookupInstance(key_holder *%s) dds.InstanceHandle_t {\n", t.Name, t.Name)
		fmt.Printf("\treturn dr.r.LookupInstance(key_holder)\n")
		fmt.Printf("}\n")

		fmt.Printf("\n\n")
	}
	fmt.Printf("\n\n")

	for _, t := range m.Modules {
		generateModule(t)
	}
}
