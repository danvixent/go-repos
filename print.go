package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

//printHelp prints help content to os.Stdout
func printHelp() {
	fmt.Print("go-repos is a tool for searching a user's GitHub Repositories\n",
		"Usage:\n", "\tgo-repos danvixent -must -name go-repos -lang go -date 2020 -desc CLI -stars 0\n\n")

	//cmds contains all supported flags and arguments
	cmds := [...]string{"danvixent", "-must", "-name", "-lang", "-date", "-desc", "-stars"}

	//usages maps cmds elements to their respective usage, for ease of writing to tw
	usages := make(map[int]string)
	usages[0] = "Username to search GitHub for"
	usages[1] = "Print only Results that match all criteria, if absent, Repositories matching at least one criteria will be displayed"
	usages[2] = "Name Of Repository to search for"
	usages[3] = "Language Base of Repository to Search for"
	usages[4] = "Creation Date Of Repository to Search For"
	usages[5] = "Repository Description to Search For"
	usages[6] = "Number Of Respository Stars to search For"

	const format = "\t%v\t%v\t\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Command", "Usage")
	fmt.Fprintf(tw, format, "-----", "------")
	for i, cmd := range cmds {
		fmt.Fprintf(tw, format, cmd, usages[i])
	}
	tw.Flush()
}

//Fprint writes the content of its receiver value to w
func (r *Result) Fprint(w io.Writer) {
	//ch allows for goroutine co-ordination when sorting results
	var ch = make(chan struct{}, 1)
	var buf = new(bytes.Buffer)

	//tab printing format
	const format = "%v\t%v\t%v\t%v\t%v\t\n"

	//tw aids printing to w
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Respository Name", "Description", "Stars", "Language", "Creation Date")
	fmt.Fprintf(tw, format, "-----", "------", "------", "------", "------")

	if *searcher.stars != 0 {
		go sorter("s", ch)
	} else {
		go sorter("", ch)
	}

	if searcher.search {
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d matching Repositories:\n\n", usr, results.Count()))
	} else {
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d  Repositories:\n\n", usr, results.Count()))
	}
	<-ch //receive from ch to be sure that sorting has finished
	for _, result := range results {
		fmt.Fprintf(tw, format, result.FullName, result.Description, result.Stars, result.Language, result.CreatedAt)
	}
	buf.WriteTo(w)
	tw.Flush()
}
