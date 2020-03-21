package main

import (
	"errors"
	"fmt"
	"sync"
)

//URL is the github api request url
const URL = "https://api.github.com/search/repositories?q=user:@&per_page=100"

var wait sync.WaitGroup
var resp = &GitResponse{}

//flag variables
var name = NameFlag("-name", "", "Search Repo Name")
var desc = DescFlag("-desc", "", "Search Repo Description")
var creationDate = DateFlag("-name", "", "Search Repo Creation Date")
var lang = LangFlag("-lang", "", "Search Repo Name")

type result []Item

func (r *result) add(item Item) error {
	if r != nil && *r != nil {
		*r = append(*r, item)
		return nil
	}
	return errors.New(" :add(item Item) called with a nil result value")
}

func (r *result) count(item Item) int {
	return len(*r)
}

// Item is the single repository data structure consisting of **only the data needed**
type Item struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Language    string `json:"language"`
}

// GitResponse contains the GitHub API response
type GitResponse struct {
	sync.Mutex        //Multiple goroutines will access the Items field
	Count      int    `json:"total_count"`
	Username   string `json:"username"`
	Items      []Item `json:"items"`
}

//program flags
type base struct { //value allows us to avoid declaring redundant Set() methods for each search struct
	val string
}

func (b *base) Set(s string) error {
	b.val = s
	return nil
}

//search by name
type searchName struct {
	base
}

//search by description
type searchDesc struct {
	base
}

//search by creation date
type searchCreationDate struct {
	base
}

//search language base
type searchLang struct {
	base
}

//SearchMux multiplexes search mechanisms
type SearchMux struct {
	toSearch bool
}

func (sn *searchName) String() string {
	return fmt.Sprintf("Name Search => %s", sn.val)
}

func (sn *searchDesc) String() string {
	return fmt.Sprintf("Description Search => %s", sn.val)
}

func (sn *searchCreationDate) String() string {
	return fmt.Sprintf("CreationDate Search => %s", sn.val)
}

func (sn *searchLang) String() string {
	return fmt.Sprintf("Language Base Search => %s", sn.val)
}
