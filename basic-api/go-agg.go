/*
This is basic example of how to use go for making API service.
Here using concurency for read XML thread. Speed is really fast (if compare with PHP or Node.js).
Special thanks to `sentdex` from youtube who made screencast about this all.
*/
package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

var wg sync.WaitGroup

// SitemapIndex is a represantation of sitemap XML file.
type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

// News is representation of news.
type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

// NewsMap is representation of news instance inside whole XML.
type NewsMap struct {
	Keywords string
	Location string
}

// NewsAggPage is representaion of struct for output page.
type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

/*
Function for handling homepage. This functions needs for creating
paths to API.

@param w - is the response writer and it write data into frame.
@param r - is pointer to request instance, it get query string from browser.
*/
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Whoa, Go is running!</h1>")
}

/*
Go routine for parsing XML file.
This is first stage of parsing.

@param c - is a chanel for buffering XML subpath.
@param Locations - is a subpath location.
*/
func newsRoutine(c chan News, Location string) {
	defer wg.Done()

	var n News
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)

	xml.Unmarshal(bytes, &n)

	resp.Body.Close()

	c <- n
}

/*
This is a second stage of XML parsing.
Here we get XML path's of news instances.
We used goroutines for creating multi-threads for load news instances
has been got from first stage of XML parsing.
And at the end we render HTML page.

@param w - is the response writer and it write data into frame.
@param r - is pointer to request instance, it get query string from browser.
*/
func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s SitemapIndex
	newsMap := make(map[string]NewsMap)
	queue := make(chan News, 30)

	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	resp.Body.Close()

	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}

	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx := range elem.Titles {
			newsMap[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "News Aggregator Page", News: newsMap}
	t, _ := template.ParseFiles("newsaggregator.html")

	t.Execute(w, p)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/agg", newsAggHandler)
	http.ListenAndServe(":8080", nil)
}
