package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"text/tabwriter"
	"time"
)

//decodePage gets url and decodes it
func decodePage(url string) {
	defer runtime.Goexit()

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("error %s: getting page %s", err, url)
		return
	}
	tmp := &GitResponse{}
	if err = json.NewDecoder(res.Body).Decode(tmp); err == nil {
		filter(tmp.Items, &results, *searcher.search) //filter tmp.Items into results
		errchan <- nil                                //send nil to errchan
		return
	}
	errchan <- fmt.Errorf("error %v: decoding page %s", err, url) //send formatted error to errchan
}

//filter decides which item in items goes into results.
//if an item is chosen, its date is formatted before it's added to results
func filter(items []Item, results *Result, search bool) {
	if search {
		for ix := range items {
			if searcher.Match(&items[ix]) {
				items[ix].fmtDate()
				mu.Lock() //necessary because of concurrency
				results.add(&items[ix])
				mu.Unlock()
			}
		}
		return
	}
	for ix := range items {
		items[ix].fmtDate()
		mu.Lock() //necessary because of concurrency
		results.add(&items[ix])
		mu.Unlock()
	}
}

//fmtDate parses and formats the CreatedAt field of i
func (i *Item) fmtDate() {
	if f, err := time.Parse(time.RFC3339, i.CreatedAt); err == nil {
		ref := f.Month().String() + " " + strconv.Itoa(f.Day()) + ", " + strconv.Itoa(f.Year()) + " " +
			strconv.Itoa(f.Hour()) + ":" + strconv.Itoa(f.Minute()) + ":" + strconv.Itoa(f.Second())
		i.CreatedAt = ref
	} else {
		fmt.Printf("Couldn't format repository %s's creation date: %v", i.FullName, err)
	}
}

//printHelp prints help content to os.Stdout
func printHelp() {
	fmt.Print("Usage:\n", "\tgo-repos danvixent -search -must -name=go-repos -lang=go -date=2020-02-11 -desc=CLI\n\n")
	//mapper maps cmds elements to their respective usage, for ease of writing to tw
	mapper := make(map[int]string)

	//cmds contains all supported flags and arguments
	cmds := [...]string{"danvixent", "-search", "-must", "-name", "-lang", "-date", "-desc"}

	mapper[0] = "Username to search GitHub for"
	mapper[1] = "Search will be done, if absent, all repositories of the user will be displayed"
	mapper[2] = "Print only Results that match all criteria, if absent, Repositories matching at least one criteria will be displayed"
	mapper[3] = "Name Of Repository to search for"
	mapper[4] = "Language Base of Repository to Search for"
	mapper[5] = "Creation Date Of Repository to Search For"
	mapper[6] = "Repository Description to Search For"

	const format = "\t%v\t%v\t\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Command", "Usage")
	fmt.Fprintf(tw, format, "-----", "------")
	for i, cmd := range cmds {
		fmt.Fprintf(tw, format, cmd, mapper[i])
	}
	tw.Flush()
}

//Gets the username command line argument.
//If not specified, "" is returned
func getUsr() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
}
