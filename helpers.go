package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

//decodePage requests url and filters it into results, any error that occurs will be sent to the errchan channel.
func decodePage(url string, page int) {
	tokens <- struct{}{}        //aquire a token
	defer func() { <-tokens }() //release token
	defer runtime.Goexit()
	defer w.Done()

	url = url + "&page=" + strconv.Itoa(page) //append the page query to the url

	res, err := http.Get(url)
	if err != nil {
		errchan <- fmt.Errorf("error getting data in page %d", page) //send formatted error to errchan
		return
	}
	res.Body.Close() //avoid resource leak

	tmp := &GitResponse{}
	if err = json.NewDecoder(res.Body).Decode(tmp); err == nil {
		filter(tmp.Items, &results, searcher.search) //filter tmp.Items into results
		errchan <- nil                               //send nil to errchan
		return
	}
	errchan <- fmt.Errorf("couldn't decode data on page %d properly", page) //send formatted error to errchan
}

//filter decides which item in items goes into results, if search is true
//else,it just copies items into results
//if an item is chosen, its date is formatted before it's added to results
func filter(items []Item, results *Result, search bool) {
	if search {
		for ix := range items {
			if searcher.Match(&items[ix]) {
				items[ix].fmtDate()
				mu.Lock() //necessary because of concurrency
				results.add(&items[ix])
				mu.Unlock()
			}
		}
		return
	}
	for ix := range items {
		items[ix].fmtDate()
		mu.Lock() //necessary because of concurrency
		results.add(&items[ix])
		mu.Unlock()
	}
}

//fmtDate parses and formats the CreatedAt field of i, any error that occurs will be sent to the errchan channel.
func (i *Item) fmtDate() {
	if f, err := time.Parse(time.RFC3339, i.CreatedAt); err == nil {
		ref := f.Month().String() + " " + strconv.Itoa(f.Day()) + ", " + strconv.Itoa(f.Year()) + " " +
			strconv.Itoa(f.Hour()) + ":" + strconv.Itoa(f.Minute()) + ":" + strconv.Itoa(f.Second())

		i.CreatedAt = ref
	} else {
		fmt.Printf("Couldn't format repository %s's creation date: %v\n", i.FullName, err)
	}
}

//Gets the username command line argument.
//If not specified, "" is returned
func getUsr() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
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
		case "stars":
			mapper[flag] = item.Stars >= *s.stars
		}
	}
	return mapper //maps are natural references
}

func sorter(field string, ch chan<- struct{}) {
	defer func() { ch <- struct{}{} }()
	switch field {
	case "s":
		sort.SliceStable(results, func(i, j int) bool { //sort results by stars
			return results[i].Stars > results[j].Stars
		})
	default:
		sort.SliceStable(results, func(i, j int) bool { //sort results by name
			return strings.ToLower(results[i].FullName) < strings.ToLower(results[j].FullName)
		})
	}
}
