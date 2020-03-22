package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"
)

//decodePage gets the page and decodes it into resp
func decodePage(url string) {
	defer wait.Done()
	defer runtime.Goexit()

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	tmp := GitResponse{}
	if err = json.NewDecoder(res.Body).Decode(&tmp); err == nil {
		if *searcher.search {
			for ix := range tmp.Items {
				mu.Lock() //necessary because of concurrency
				if searcher.check(&(tmp.Items[ix])) {
					results.add(tmp.Items[ix])
				}
				mu.Unlock()
			}
			return
		}

		for ix := range tmp.Items {
			resp.Lock() //necessary because of concurrency
			resp.add(tmp.Items[ix])
			resp.Unlock()
		}
		return
	}
	fmt.Printf("error %s: decoding page %s", err, url)
}

// printResults() sends the data to the client using a buffer
func printResults() {
	var sorte sync.WaitGroup
	buf := new(bytes.Buffer)
	const format = "%v\t%v\t%v\t%v\t\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, '	', 0)

	fmt.Fprintf(tw, format, "Respository Name", "Description", "Language", "Creation Date")
	fmt.Fprintf(tw, format, "-----", "------", "------", "------")
	if *searcher.search {
		{
			go sorter(results, &sorte)
			sorte.Add(1)
			sorte.Wait()
		}
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d matching Repositories:", usr, results.Count()))
		for _, result := range results {
			fmt.Fprintf(tw, format, result.FullName, result.Description, result.Language, result.CreatedAt)
		}
	} else {
		{
			go sorter(resp.Items, &sorte)
			sorte.Add(1)
			sorte.Wait()
		}
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d Repositories", usr, resp.RepoCount()))
		for _, item := range resp.Items {
			fmt.Fprintf(tw, format, item.FullName, item.Description, item.Language, item.CreatedAt)
		}
	}
	buf.WriteTo(os.Stdout)
	tw.Flush()
}

func fmtDates(items []Item) {
	for i := range items {
		r := &(items[i])
		if format, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
			ref := strconv.Itoa(format.Year()) + " " + format.Month().String() + " " +
				strconv.Itoa(format.Hour()) + ":" + strconv.Itoa(format.Minute()) + ":" + strconv.Itoa(format.Second())
			r.CreatedAt = ref
		} else {
			log.Fatal(err)
		}
	}
}

func printHelp() {
	fmt.Println("go-repos help:")
	fmt.Println("Usage Example:", "go-repos danvixent -search -must -name=go-repos -lang=go -date=2020-02-11 -desc=CLI")
	mapper := make(map[int]string)
	cmds := [...]string{"danvixent", "-search", "-must", "-name", "-lang", "-date", "-desc"}

	mapper[0] = "Username to search GitHub for"
	mapper[1] = "Search will be done, if absent, all repositories of the user will be displayed"
	mapper[2] = "Print only Results that match all criteria, if absent, Repositories matching at least one criteria will be displayed"
	mapper[3] = "Name Of Repository to search for"
	mapper[4] = "Language Base of Repository to Search for"
	mapper[5] = "Creation Date Of Repository to Search For"
	mapper[6] = "Repository Description to Search For"

	const format = "%v\t%v\t\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, '	', 0)
	fmt.Fprintf(tw, format, "Command", "Usage")
	fmt.Fprintf(tw, format, "-----", "------")
	for i, cmd := range cmds {
		fmt.Fprintf(tw, format, cmd, mapper[i])
	}
	tw.Flush()
}

func sorter(items []Item, waiter *sync.WaitGroup) {
	defer waiter.Done()
	defer runtime.Goexit()
	sort.SliceStable(items, func(i, j int) bool { //sort the repos by name
		return items[i].FullName < items[j].FullName
	})
}

func getUsr() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
}
