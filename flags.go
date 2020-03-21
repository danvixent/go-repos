package main

import (
	"flag"
)

//NameFlag defines a name flag variable,with the flag name,value and usage.
func NameFlag(name string, value string, usage string) *string {
	f := searchName{base{value}}
	flag.CommandLine.Var(&f, name, usage)
	return &f.val
}

//DescFlag defines a name flag variable,with the flag name,value and usage.
func DescFlag(name string, value string, usage string) *string {
	f := searchDesc{base{value}}
	flag.CommandLine.Var(&f, name, usage)
	return &f.val
}

//DateFlag defines a name flag variable,with the flag name,value and usage.
func DateFlag(name string, value string, usage string) *string {
	f := searchCreationDate{base{value}}
	flag.CommandLine.Var(&f, name, usage)
	return &f.val
}

//LangFlag defines a name flag variable,with the flag name,value and usage.
func LangFlag(name string, value string, usage string) *string {
	f := searchLang{base{value}}
	flag.CommandLine.Var(&f, name, usage)
	return &f.val
}
