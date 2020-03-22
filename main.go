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

func main() {
	flag.Parse()
	if usr == "" || *help {
		printHelp()
		os.Exit(3)
	}
	err := Fetch()
	if err != nil {
		log.Fatal(fmt.Errorf("error fetching data: %v", err))
	}

	if *searcher.search {
		fmtDates(results)
	} else {
		fmtDates(resp.Items)
	}
	printResults()
}

// Fetch fetches the data from GitHub and paginates if neccessary
func Fetch() error {
	url := strings.Replace(URL, "@", usr, 1)

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf(" :error making first request: %v", err)
	}

	if err = json.NewDecoder(res.Body).Decode(resp); err == nil {
		if resp.RepoCount() > 100 {
			Paginate(url)
		}
		resp.Username = usr
		return nil
	}
	return fmt.Errorf(" :couldn't decode page into resp: %v", err)
}

//Paginate goes through each page up till the end
func Paginate(url string) {
	num := (resp.RepoCount() / 100) //compute number pages to request for
	if resp.RepoCount()%100 != 0 {  //if the number of repos isn't a multiple of 100, one more page will be needed
		num++
	}
	//spawn new goroutines to decode each page
	for i := 2; i <= num; i++ {
		url = url + "&page=" + strconv.Itoa(i)
		fmt.Println("Gotten page", i, "url=", url)
		go decodePage(url)
		wait.Add(1)
	}
	wait.Wait()
}
