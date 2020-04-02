package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
)

//URL is the base github API request url
const URL = "https://api.github.com/search/repositories?q=user:@&per_page=100"

var (
	//mu locks results for concurrent access during filtering
	mu sync.Mutex

	//tokens is a counting semaphore,helps limit the number of active decodePage() goroutines
	tokens chan struct{}

	//errchan receives errors sent during decoding multiple pages by goroutines
	errchan = make(chan error, 2)

	//w waits for all goroutines using errchan to finish
	w sync.WaitGroup

	//initResp holds the first decoded response from the github API
	initResp = &GitResponse{}

	//searcher holds all search-related flag variables.
	searcher = Search{
		//flag variables
		search: flags.Bool("search", false, "To search Repo Data"),
		name:   NameFlag("name", "", "Search Repo Name"),
		desc:   DescFlag("desc", "", "Search Repo Description"),
		date:   DateFlag("date", "", "Search Repo Creation Date"),
		lang:   LangFlag("lang", "", "Search Repo Name"),
		must:   flags.Bool("must", false, "Must match all criteria"),
	}

	//help is the help flag variable.
	help = flags.Bool("help", false, "Help")

	//flags is the user defined flagset.
	flags = flag.NewFlagSet("flags", flag.ExitOnError)

	//SetFlags holds the names of the flags that were set.
	SetFlags []string

	//results holds the result of entire operation
	//and is printed at the end of the operation,not initResp.
	results = make(Result, 0)

	//usr is the username command line argument.
	usr string
)

type (
	//Result is a named type of []Item
	Result []Item

	// Item is the single repository data structure consisting of **only the data needed**
	Item struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		Language    string `json:"language"`
	}

	// GitResponse contains the GitHub API response
	GitResponse struct {
		Count int    `json:"total_count"`
		Items []Item `json:"items"`
	}

	base struct { //base allows us to avoid declaring redundant Set() & String() methods for each search struct
		val string
	}

	//search by name
	searchName struct {
		base
	}

	//search by description
	searchDesc struct {
		base
	}

	//search by creation date
	searchCreationDate struct {
		base
	}

	//search language base
	searchLang struct {
		base
	}

	//Search multiplexes search mechanisms
	Search struct {
		//search flag
		search *bool

		//name flag
		name *string

		//description flag
		desc *string

		//date flag
		date *string

		//language base flag
		lang *string

		//must flag
		must *bool
	}
)

//add() adds item to this Result
func (r *Result) add(item *Item) {
	*r = append(*r, *item)
}

//Count returns the length of this Result
func (r *Result) Count() int {
	return len(*r)
}

//RepoCount returns the number of repositories
func (g *GitResponse) RepoCount() int {
	return len(g.Items)
}

//Set() sets the value of the val field to s
func (b *base) Set(s string) error {
	b.val = s
	return nil
}

//String returns the val field value
func (b *base) String() string {
	return b.val
}

//Match finds out if item matches the search criteria
func (s *Search) Match(item *Item) bool {
	mapper := flagMapper(item, s)
	if !*s.must { //if the must flag is false
		for _, val := range mapper {
			if val { //and at least one condition is met return true
				return true
			}
		}
		return false
	}

	for _, val := range mapper {
		if !val { //if at least one condition is false return false
			return false
		}
	}
	return true
}

//flagMapper returns a map containing each set flag and a corresponding boolean value
//representing if the item meets the flag conditions
func flagMapper(item *Item, s *Search) map[string]bool {
	mapper := make(map[string]bool)
	for _, flag := range SetFlags {
		switch flag {
		case "name":
			mapper[flag] = strings.Contains(strings.ToLower(item.FullName), strings.ToLower(*s.name))
		case "desc":
			mapper[flag] = strings.Contains(strings.ToLower(item.Description), strings.ToLower(*s.desc))
		case "date":
			mapper[flag] = strings.Contains(strings.ToLower(item.CreatedAt), strings.ToLower(*s.date))
		case "lang":
			mapper[flag] = strings.Contains(strings.ToLower(item.Language), strings.ToLower(*s.lang))
		}
	}
	return mapper //maps are natural references
}

//Fprint writes the content of its receiver value to w
func (r *Result) Fprint(w io.Writer) {
	//ch allows for goroutine co-ordination when sorting results
	var ch = make(chan struct{}, 1)
	var buf = new(bytes.Buffer)

	//tab printing format
	const format = "%v\t%v\t%v\t%v\t\n"

	//tw aids printing to w
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)

	fmt.Fprintf(tw, format, "Respository Name", "Description", "Language", "Creation Date")
	fmt.Fprintf(tw, format, "-----", "------", "------", "------")

	go func() {
		sort.SliceStable(results, func(i, j int) bool { //sort results by name
			return strings.ToLower(results[i].FullName) < strings.ToLower(results[j].FullName)
		})
		//when finished send an empty struct. Note that ch is buffered
		ch <- struct{}{}
		close(ch) //close ch...just a precaution
	}()

	if *searcher.search {
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d matching Repositories:\n\n", usr, results.Count()))
	} else {
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d  Repositories:\n\n", usr, results.Count()))
	}
	<-ch //receive from ch to be sure that sorting has finished
	for _, result := range results {
		fmt.Fprintf(tw, format, result.FullName, result.Description, result.Language, result.CreatedAt)
	}
	buf.WriteTo(w)
	tw.Flush()
}
