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
	wait     sync.WaitGroup
	mu       sync.Mutex
	resp     = &GitResponse{}
	searcher = Search{
		//flag variables
		search: flag.Bool("search", false, "To search Repo Data"),
		name:   NameFlag("name", "", "Search Repo Name"),
		desc:   DescFlag("desc", "", "Search Repo Description"),
		date:   DateFlag("date", "", "Search Repo Creation Date"),
		lang:   LangFlag("lang", "", "Search Repo Name"),
		must:   flag.Bool("must", false, "Must match all criteria"),
	}
	help    = flag.Bool("help", false, "Help")
	results = make(Result, 0)
	usr     = getUsr()
)

type (
	//Result is used to store the results if a search operation is performed
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

	base struct { //value allows us to avoid declaring redundant Set() & String() methods for each search struct
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
		search *bool
		name   *string
		desc   *string
		date   *string
		lang   *string
		must   *bool
	}
)

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

func (b *base) Set(s string) error {
	b.val = s
	return nil
}

//String returns the  Value
func (b *base) String() string {
	return b.val
}

func (s *Search) check(item *Item) bool {
	nonempty := []string{}

	if *s.name != "" {
		nonempty = append(nonempty, "name")
	}

	if *s.desc != "" {
		nonempty = append(nonempty, "desc")
	}
	if *s.date != "" {
		nonempty = append(nonempty, "date")
	}

	if *s.lang != "" {
		nonempty = append(nonempty, "lang")
	}

	if !*s.must {
		for _, val := range nonempty {
			switch val {
			case "name":
				if *s.name == strings.SplitAfter((item.FullName), "/")[1] {
					return true
				}
			case "desc":
				if strings.Contains(strings.ToLower(item.Description), *s.desc) {
					return true
				}
			case "date":
				if strings.Contains(strings.ToLower(item.CreatedAt), *s.date) {
					return true
				}
			case "lang":
				if *s.lang == item.Language {
					return true
				}
			}
		}
		return false
	}
	bools := make([]bool, 0, len(nonempty))
	for _, val := range nonempty {
		switch val {
		case "name":
			if strings.Contains(strings.ToLower(item.FullName), *s.name) {
				bools = append(bools, true)
				continue
			}
			bools = append(bools, false)
		case "desc":
			if strings.Contains(strings.ToLower(item.Description), *s.desc) {
				bools = append(bools, true)
				continue
			}
			bools = append(bools, false)
		case "date":
			if strings.Contains(strings.ToLower(item.CreatedAt), *s.date) {
				bools = append(bools, true)
				continue
			}
			bools = append(bools, false)
		case "lang":
			if strings.Contains(strings.ToLower(item.Language), *s.lang) {
				bools = append(bools, true)
				continue
			}
			bools = append(bools, false)
		}
	}
	for _, b := range bools {
		if !b {
			return false
		}
	}
	return true
}

//Fprint writes the content of its receiver value to w
func (r *Result) Fprint(w io.Writer) {
	var ch = make(chan struct{}, 1)
	var buf = new(bytes.Buffer)
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)

	fmt.Fprintf(tw, format, "Respository Name", "Description", "Language", "Creation Date")
	fmt.Fprintf(tw, format, "-----", "------", "------", "------")

	go func() {
		sort.SliceStable(results, func(i, j int) bool { //sort the repos by name
			return results[i].FullName < results[j].FullName
		})
		ch <- struct{}{}
		close(ch)
	}()

	if *searcher.search {
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d matching Repositories:\n\n", usr, results.Count()))
	} else {
		buf.WriteString(fmt.Sprintf("GitHub User %s has %d  Repositories:\n\n", usr, results.Count()))
	}
	<-ch
	for _, result := range results {
		fmt.Fprintf(tw, format, result.FullName, result.Description, result.Language, result.CreatedAt)
	}
	buf.WriteTo(w)
	tw.Flush()
}
