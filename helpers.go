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

//decodePage requests url and filters it into results, any error that occurs will be sent to the errchan channel.
func decodePage(url string) {
	w.Add(1)
	defer w.Done()
	tokens <- struct{}{}        //aquire a token
	defer func() { <-tokens }() //release a token
	defer runtime.Goexit()

	res, err := http.Get(url)
	if err != nil {
		errchan <- fmt.Errorf("error %s: getting page %s", err, url) //send formatted error to errchan
		return
	}
	defer res.Body.Close() //avoid resource leak

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

//fmtDate parses and formats the CreatedAt field of i, any error that occurs will be sent to the errchan channel.
func (i *Item) fmtDate() {
	if f, err := time.Parse(time.RFC3339, i.CreatedAt); err == nil {
		ref := f.Month().String() + " " + strconv.Itoa(f.Day()) + ", " + strconv.Itoa(f.Year()) + " " +
			strconv.Itoa(f.Hour()) + ":" + strconv.Itoa(f.Minute()) + ":" + strconv.Itoa(f.Second())

		i.CreatedAt = ref
	} else {
		fmt.Printf("Couldn't format repository %s's creation date: %v\n", i.FullName, err)
	}
}

//printHelp prints help content to os.Stdout
func printHelp() {
	fmt.Print("go-repos is a tool for searching a user's GitHub Repositories\n",
		"Usage:\n", "\tgo-repos danvixent -search -must -name go-repos -lang go -date 2020 -desc CLI\n\n")

	//cmds contains all supported flags and arguments
	cmds := [...]string{"danvixent", "-search", "-must", "-name", "-lang", "-date", "-desc"}

	//usages maps cmds elements to their respective usage, for ease of writing to tw
	usages := make(map[int]string)
	usages[0] = "Username to search GitHub for"
	usages[1] = "Search will be done, if absent, all Repositories of the user will be displayed"
	usages[2] = "Print only Results that match all criteria, if absent, Repositories matching at least one criteria will be displayed"
	usages[3] = "Name Of Repository to search for"
	usages[4] = "Language Base of Repository to Search for"
	usages[5] = "Creation Date Of Repository to Search For"
	usages[6] = "Repository Description to Search For"

	const format = "\t%v\t%v\t\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Command", "Usage")
	fmt.Fprintf(tw, format, "-----", "------")
	for i, cmd := range cmds {
		fmt.Fprintf(tw, format, cmd, usages[i])
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
