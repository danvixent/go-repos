package main

import (
	"errors"
	"flag"
	"strings"
	"sync"
)

//URL is the github api request url
const URL = "https://api.github.com/search/repositories?q=user:@&per_page=100"

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
		must:   flag.Bool("m", false, "Must match all criteria"),
	}
	help    = flag.Bool("h", false, "Help")
	results = make(result, 0)
	usr     = flag.Arg(1)
)

type (
	result []Item

	// Item is the single repository data structure consisting of **only the data needed**
	Item struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		Language    string `json:"language"`
	}

	// GitResponse contains the GitHub API response
	GitResponse struct {
		sync.Mutex //Multiple goroutines will access the Items field
		Username   string
		Count      int    `json:"total_count"`
		Items      []Item `json:"items"`
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

func (r *result) add(item Item) error {
	if r != nil && *r != nil {
		*r = append(*r, item)
		return nil
	}
	return errors.New(" :add(item Item) called with a nil result value")
}

func (r *result) Count() int {
	return len(*r)
}

//RepoCount returns the number of repositories
func (g *GitResponse) RepoCount() int {
	return len(g.Items)
}

func (g *GitResponse) add(item Item) error {
	g.Items = append(g.Items, item)
	return nil
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

	if *s.name != "" && strings.Contains(strings.ToLower(item.FullName), *s.name) &&
		*s.desc != "" && strings.Contains(strings.ToLower(item.Description), *s.desc) &&
		*s.date != "" && strings.Contains(strings.ToLower(item.CreatedAt), *s.date) {
		return true
	}

	return false
}
