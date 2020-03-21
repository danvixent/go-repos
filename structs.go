package main

import "sync"

//URL is the github api request url
const URL = "https://api.github.com/search/repositories?q=user:@&per_page=100"

var wait sync.WaitGroup
var resp = &GitResponse{}

// Item is the single repository data structure consisting of **only the data needed**
type Item struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// GitResponse contains the GitHub API response
type GitResponse struct {
	sync.Mutex        //Multiple goroutines will access the Items field
	Count      int    `json:"total_count"`
	Username   string `json:"username"`
	Items      []Item `json:"items"`
}
