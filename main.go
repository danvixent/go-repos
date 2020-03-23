package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func init() {
	if len(os.Args) >= 3 {
		flagger.Parse(os.Args[2:])
	}

	usr = getUsr()
	if usr == "" || *help {
		printHelp()
		os.Exit(3)
	}
}

func main() {
	fmt.Println("search =", *searcher.search)
	fmt.Println("lang =", *searcher.lang)
	fmt.Println("date =", *searcher.date)
	fmt.Println("desc =", *searcher.desc)
	fmt.Println("must =", *searcher.must)
	fmt.Println("usr =", usr)

	err := Fetch()
	if err != nil {
		log.Fatal(fmt.Errorf("error fetching data: %v", err))
	}
	results.Fprint(os.Stdout)
}

// Fetch fetches the data from GitHub and paginates if neccessary
func Fetch() error {
	url := strings.Replace(URL, "@", usr, 1)

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making first request: %v", err)
	}

	if err = json.NewDecoder(res.Body).Decode(resp); err == nil {
		filter(resp.Items, *searcher.search)
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
