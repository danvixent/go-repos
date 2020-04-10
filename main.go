package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

//initialize username argument and flag variables
func init() {
	usr = getUsr()
	if len(os.Args) >= 3 {
		//work around for the username argument, start parsing flags from the third cmd line argument
		//note that all flags are themselves cmd line arguments
		flags.Parse(os.Args[2:])
	}

	//visit all flags that were set at the cmd line and append their names to SetFlags
	flags.Visit(func(f *flag.Flag) {
		SetFlags = append(SetFlags, f.Name)
		if !searcher.search { //set search to true, since a search will be performed
			searcher.search = true
		}
	})
}

func main() {
	if usr == "" || *help {
		printHelp()
		os.Exit(3)
	}

	if '-' == usr[0] { //check if the username argument has an '-' character at its beginning
		fmt.Print("Warning: username argument has a '-' character at its beginning")
		os.Exit(3)
	}

	err := Fetch()
	if err != nil {
		log.Fatal(fmt.Errorf("error fetching data: %v", err))
	}
	results.Fprint(os.Stdout)
}

// Fetch fetches the data from GitHub and paginates,if neccessary.
func Fetch() error {
	//Replace '@' with the given username
	url := strings.Replace(URL, "@", usr, 1)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("check your internet connection")
	}

	if err = json.NewDecoder(res.Body).Decode(initResp); err == nil {
		filter(initResp.Items, &results, searcher.search) //filter the initial resp.Items into results
		if initResp.RepoCount() > 100 {
			Paginate(url)
		}
		return nil
	}
	return fmt.Errorf("couldn't decode data properly")
}

//Paginate goes through each page up till the end
func Paginate(url string) {
	numPages := (initResp.RepoCount() / 100) //compute number pages to request for
	if initResp.RepoCount()%100 != 0 {       //if the number of repos isn't a multiple of 100, one more page will be needed
		numPages++
	}
	//initialize the tokens channel
	tokens = make(chan struct{}, 19)
	//spawn new goroutines to decode each page, starting from the second; Fetch() already got the first page
	for i := 2; i <= numPages; i++ {
		w.Add(1)
		go decodePage(url, i)
	}

	//closer
	go func() {
		w.Wait()
		close(errchan)
	}()

	//begin receiving errors from errchan, report any non-nil error
	for err := range errchan {
		if err != nil {
			fmt.Printf("%v", err)
		}
	}
}
