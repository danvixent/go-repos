package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//intialize flag variables and username argument
func init() {
	if len(os.Args) >= 3 {
		flagger.Parse(os.Args[2:])
	}

	//append the name of the flags that were set to SetFlags
	flagger.Visit(func(f *flag.Flag) {
		SetFlags = append(SetFlags, f.Name)
	})
	usr = getUsr()
}

func main() {
	//if username argument is missing or the help flag is specified call printhelp()
	if usr == "" || *help {
		printHelp()
		os.Exit(3)
	}

	err := Fetch()
	if err != nil {
		log.Fatal(fmt.Errorf("error fetching data: %v", err))
	}
	results.Fprint(os.Stdout)
}

// Fetch fetches the data from GitHub and paginates if neccessary
func Fetch() error {
	//Here, replace '@' with the given username
	url := strings.Replace(URL, "@", usr, 1)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making first request: %v", err)
	}

	if err = json.NewDecoder(res.Body).Decode(resp); err == nil {
		filter(resp.Items, &results, *searcher.search) //filter the initial resp.Items into results
		if resp.RepoCount() > 100 {
			Paginate(url)
		}
		resp.Username = usr
		return nil
	}
	return fmt.Errorf("couldn't decode page into resp: %v", err)
}

//Paginate goes through each page up till the end
func Paginate(url string) {
	numPages := (resp.RepoCount() / 100) //compute number pages to request for
	if resp.RepoCount()%100 != 0 {       //if the number of repos isn't a multiple of 100, one more page will be needed
		numPages++
	}
	//spawn new goroutines to decode each page, starting from the second; Fetch() already got the first page
	for i := 2; i <= numPages; i++ {
		url = url + "&page=" + strconv.Itoa(i) //append the page query to the url
		go decodePage(url)
	}
	//begin receiving errors from errchan, report if not nil
	for i := numPages - 1; i > 0; i-- {
		if err := <-errchan; err != nil {
			fmt.Printf("error while paginating: %v", err)
		}
	}
	close(errchan) //close errchan
}
