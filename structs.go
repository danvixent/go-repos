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

//URL is the github api request url
const (
	URL    = "https://api.github.com/search/repositories?q=user:@&per_page=100"
	format = "%v\t%v\t%v\t%v\t\n"
)

var (
	mu       sync.Mutex
	errchan  = make(chan error, 1)
	resp     = &GitResponse{}
	searcher = Search{
		//flag variables
		search: flagger.Bool("search", false, "To search Repo Data"),
		name:   NameFlag("name", "", "Search Repo Name"),
		desc:   DescFlag("desc", "", "Search Repo Description"),
		date:   DateFlag("date", "", "Search Repo Creation Date"),
		lang:   LangFlag("lang", "", "Search Repo Name"),
		must:   flagger.Bool("must", false, "Must match all criteria"),
	}
	help    = flagger.Bool("help", false, "Help")
	flagger = flag.NewFlagSet("flagger", flag.ContinueOnError)

	//SetFlags holds the names of the flags that were set
	SetFlags = make([]string, 0)

	results = make(Result, 0) //holds results of entire operation, will change in size a lot
	usr     string            //username command line argument
)

type (
	//Result is used to store the results
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
		Username string
		Count    int    `json:"total_count"`
		Items    []Item `json:"items"`
	}

	base struct { //base allows us to avoid declaring redundant Set() & String() methods for each search struct
		val string
	}
	//username flag
	username struct {
		base
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
func (r *Result) add(item *Item) error {
	*r = append(*r, *item)
	return nil
}

//Count returns the length of the Result
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
	mapper := s.flagMapper(item)
	if !*s.must { //if the must flag is false
		for _, val := range *mapper {
			if val { //if at least one condition is met return true
				return true
			}
		}
		return false
	}

	for _, val := range *mapper {
		if !val { //if at least one condition is met return true
			return false
		}
	}
	return true
}

func (s *Search) flagMapper(item *Item) *map[string]bool {
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
	return &mapper
}

//Fprint writes the content of its receiver value to w
func (r *Result) Fprint(w io.Writer) {
	//ch allows for goroutine co-ordination when sorting results
	var ch = make(chan struct{}, 1)
	var buf = new(bytes.Buffer)

	//tw aids printing to w
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)

	fmt.Fprintf(tw, format, "Respository Name", "Description", "Language", "Creation Date")
	fmt.Fprintf(tw, format, "-----", "------", "------", "------")

	go func() {
		sort.SliceStable(results, func(i, j int) bool { //sort results by name
			return results[i].FullName < results[j].FullName
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
