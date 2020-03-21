package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/post", fetch)
	http.ListenAndServe(":8080", nil)
}

// fetch() fetches the data from GitHub and paginates if neccessary
func fetch(w http.ResponseWriter, r *http.Request) {
	usr := r.FormValue("firstname")
	url := strings.Replace(URL, "@", usr, 1)

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	if err = json.NewDecoder(res.Body).Decode(resp); err == nil {
		if resp.Count > 100 {
			Paginate(resp, url)
		}
		fmtDates(resp)
		resp.Username = usr
		sort.SliceStable(resp.Items, func(i, j int) bool { //sort the repos by name
			return resp.Items[i].FullName < resp.Items[j].FullName
		})
		sendResp(resp, &w)
		return
	}
	log.Fatal(err)
}

func fmtDates(resp *GitResponse) {
	for i := range resp.Items {
		r := &(resp.Items[i])
		if format, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
			ref := strconv.Itoa(format.Year()) + " " + format.Month().String() + " " +
				strconv.Itoa(format.Hour()) + ":" + strconv.Itoa(format.Minute()) + ":" + strconv.Itoa(format.Second())
			r.CreatedAt = ref
		} else {
			log.Fatal(err)
		}
	}
}

//Paginate goes through each page up till the end
func Paginate(resp *GitResponse, url string) {
	num := (resp.Count / 100) //compute number pages to request for
	if resp.Count%100 != 0 {  //if the number of repos isn't a multiple of 100, one more page will be needed
		num++
	}
	//spawn new goroutines to decode each page
	for i := 2; i <= num; i++ {
		url = url + "&page=" + strconv.Itoa(i)
		fmt.Println("Gotten page", i, "url=", url)
		go decodePage(url, resp)
		wait.Add(1)
	}
	wait.Wait()
}

//decodePage gets the page and decodes it into resp
func decodePage(url string, resp *GitResponse) {
	defer wait.Done()
	defer runtime.Goexit()

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	tmp := GitResponse{}
	if err = json.NewDecoder(res.Body).Decode(&tmp); err == nil {
		for ix := range tmp.Items {
			resp.Lock() //necessary because of concurrency
			resp.Items = append(resp.Items, tmp.Items[ix])
			resp.Unlock()
		}
		return
	}
	fmt.Println(err)
}

// sendResp() sends the data to the client using a buffer
func sendResp(data *GitResponse, w *http.ResponseWriter) {
	t := template.New("Response Data")
	t, _ = t.Parse(`
    <p>
	  Github User {{ .Username }} has {{ .Count }} repositories as follows:
    </p>
    <table>
        <tr style="display: table-row">
          <th>Repository Name</th>
		  <th>Description</th>
		  <th>Created At</th>
		</tr>
		{{ with .Items }}

			{{ range . }}
        		<tr style="display: table-row">
          			<td id="fn">{{ .FullName }}</td>
		  			<td id="dsc">{{ .Description }}</td>
		  			<td id="cat">{{ .CreatedAt }}</td>
				</tr>
			{{ end }}

		{{ end }}
      </table>
	`)

	buf := new(bytes.Buffer)
	t.Execute(buf, data) //*data
	(*w).Write(buf.Bytes())
	(*w).(http.Flusher).Flush()
}
