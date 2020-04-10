package main

import (
	"flag"
	"sync"
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

	//flags is the user defined flagset.
	flags = flag.NewFlagSet("flags", flag.ExitOnError)

	//searcher holds all search-related flag variables.
	searcher = Search{
		//flag variables
		name:  flags.String("name", "", "Search Repo Name"),
		desc:  flags.String("desc", "", "Search Repo Description"),
		date:  flags.String("date", "", "Search Repo Creation Date"),
		lang:  flags.String("lang", "", "Search Repo Name"),
		stars: flags.Int("stars", 0, "Number of Stars to search for"),
		must:  flags.Bool("must", false, "Must match all criteria"),
	}

	//help is the help flag variable.
	help = flags.Bool("help", false, "Help")

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
	Result []*Item

	// Item is the single repository data structure consisting of **only the data needed**
	Item struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		Language    string `json:"language"`
		Stars       int    `json:"stargazers_count"`
	}

	// GitResponse contains the GitHub API response
	GitResponse struct {
		Count int    `json:"total_count"`
		Items []Item `json:"items"`
	}

	//Search multiplexes search mechanisms
	Search struct {
		//signal that search will be performed
		search bool

		//name flag
		name *string

		//description flag
		desc *string

		//date flag
		date *string

		//language base flag
		lang *string

		//stars flags
		stars *int

		//must flag
		must *bool
	}
)

//add() adds item to this Result
func (r *Result) add(item *Item) {
	if len(item.Description) > 70 {
		item.Description = item.Description[0:70] + "..."
	}
	*r = append(*r, item)
}

//Count returns the length of this Result
func (r Result) Count() int {
	return len(r)
}

//RepoCount returns the number of repositories
func (g *GitResponse) RepoCount() int {
	return len(g.Items)
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
