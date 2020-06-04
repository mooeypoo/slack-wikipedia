package wikipedia

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "log"
	"github.com/araddon/dateparse"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TODO: Maybe make this configurable? Allow for other languages?
var wikiBaseArticlePath = "https://en.wikipedia.org/wiki/%s"
var wikiRESTEndpoint = "https://en.wikipedia.org/api/rest_v1/"
var wikiRESTsummary = "page/summary/%s?redirect=true"
var wikiRESTrelated = "page/related/%s"
var wikiActionAPIendpoint = "https://en.wikipedia.org/w/api.php"
var wikiAnalyticsPageviewsEndpoint = "https://wikimedia.org/api/rest_v1/metrics/pageviews/top/en.wikipedia/all-access/%d/%02d/%02d" // "2020/06/02"

// Normalized struct for page response
type Page struct {
	Title   string
	Extract string
	Image   string
	URL     string
}

// Normalized struct for multiple page response
type MultiplePages struct {
	Pages []Page
}

// Normalized struct for list
type Pagelist struct {
	Pages []PagelistPage
}
type PagelistPage struct {
	Title string
	URL   string
	Rank  int
	Info  string
}

// Fetch the summary of a specific Wikipedia page given by its title
func FetchSummary(title string) (resp MultiplePages) {
	safeTitle := prepTitleForURLQuery(title)
	url := fmt.Sprintf(wikiRESTEndpoint+wikiRESTsummary, safeTitle)
	fmt.Println("Fetching summary: " + url)

	body, readErr := fetchFromApi(url)
	if readErr != nil {
		return getNotFound()
	}

	return processRESTApiResult(body, false)
}

// Fetch related pages to the given term
func FetchRelated(term string) (resp MultiplePages) {
	safeTitle := prepTitleForURLQuery(term)
	url := fmt.Sprintf(wikiRESTEndpoint+wikiRESTrelated, safeTitle)
	fmt.Println("Fetching related: " + url)

	body, readErr := fetchFromApi(url)
	if readErr != nil {
		return getNotFound()
	}

	return processRESTApiResult(body, true)
}

// Fetch search results from Wikipedia given the search string
func FetchSearch(searchString string) (resp MultiplePages) {
	safeTitle := prepTitleForURLQuery(searchString)

	params := url.Values{}

	params.Add("action", "query")
	params.Add("format", "json")
	params.Add("prop", "extracts|pageimages|info")
	params.Add("generator", "search")
	params.Add("redirects", "1")
	params.Add("exchars", "250")
	params.Add("exlimit", "5")
	params.Add("exintro", "1")
	params.Add("explaintext", "1")
	params.Add("inprop", "url")
	params.Add("gsrlimit", "5")
	params.Add("gsrwhat", "text")
	params.Add("gsrsearch", safeTitle)

	url := wikiActionAPIendpoint + "?" + params.Encode()
	fmt.Println("Fetching search: " + url)

	body, readErr := fetchFromApi(url)
	if readErr != nil {
		return getNotFound()
	}

	return processActionAPIResult(body)
}
func ParseTimeString(datestring string) (parsed time.Time) {
	datestring = strings.TrimSpace(datestring)
	fmt.Printf("Parsing time string \"%s\"", datestring)
	t := time.Now()
	if len(datestring) != 0 {
		parsedTime, err := dateparse.ParseAny(datestring)
		if err != nil {
			t = time.Now()
			fmt.Println("Failed to parse given date string.")
			fmt.Println(err)
		} else {
			t = parsedTime
		}
	}
	return t
}
func FetchTopPageviews(datestring string) (resp Pagelist, requestedTime time.Time) {
	t := ParseTimeString(datestring)
	// Build the url
	// fmt.Println("FetchTopPageviews Date: " + t.Format("January 2, 2005"))
	url := fmt.Sprintf(wikiAnalyticsPageviewsEndpoint, t.Year(), int(t.Month()), t.Day())
	fmt.Println("FetchTopPageviews URL: " + url)
	body, readErr := fetchFromApi(url)
	if readErr != nil {
		return Pagelist{
			Pages: []PagelistPage{PagelistPage{"Not found.", "", 0, ""}},
		}, t
	}

	return processAnalyticsPageviews(body), t
}

// Prepare a given string to be used in a URL query
// Outputs a URL-query safe string with %20 representing spaces
func prepTitleForURLQuery(text string) (urlSafeTitle string) {
	// Purposefully change spaces to %20 because queryEscape transforms them to +
	// And the Wikipedia REST api seems to not like that at all
	safeTitle := strings.TrimSpace(text)
	safeTitle = url.QueryEscape(safeTitle)
	safeTitle = strings.ReplaceAll(safeTitle, "+", "%20")
	return safeTitle
}

// Fetch data from the given API link
// Return the bytstream for the body of the reply to be processed
func fetchFromApi(link string) (body []byte, err error) {
	wikiClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	fakeBody := []byte{}

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return fakeBody, err
	}

	req.Header.Set("User-Agent", "slack-wikipedia-bot")

	res, getErr := wikiClient.Do(req)
	if getErr != nil {
		return fakeBody, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return ioutil.ReadAll(res.Body)
}

// Get result from the Wikipedia Action API and output a normalized
// data structure through MultiplePages
func processActionAPIResult(body []byte) (page MultiplePages) {
	record := ActionAPIGeneratorResponse{}
	jsonErr := json.Unmarshal(body, &record)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return getNotFound()
	}

	collection := []Page{}
	for _, page := range record.Query.Pages {
		collection = append(collection, Page{
			page.Title,
			strings.TrimSpace(page.Extract),
			page.Thumbnail.Source,
			page.Canonicalurl})
	}
	return MultiplePages{Pages: collection}
}

// Get result from the Wikipedia RESTBASE API and output a normalized
// data structure through MultiplePages
// isMultple parameter should be set to true if the request is expected
// to return a JSON structure that holds multiple results. False otherwise.
func processRESTApiResult(body []byte, isMultiple bool) (page MultiplePages) {
	if isMultiple {
		record := MultiplePageResponseREST{}
		jsonErr := json.Unmarshal(body, &record)
		if jsonErr != nil {
			fmt.Println(jsonErr)
			return getNotFound()
		} else {
			collection := []Page{}
			for _, page := range record.Pages {
				collection = append(collection, Page{
					page.Titles.Normalized,
					strings.TrimSpace(page.Extract),
					page.Thumbnail.Source,
					page.ContentUrls.Desktop.Page})
			}
			return MultiplePages{Pages: collection}
		}
	} else {
		record := PageResponseREST{}
		jsonErr := json.Unmarshal(body, &record)
		if jsonErr != nil {
			fmt.Println(jsonErr)
			return getNotFound()
		} else {
			return MultiplePages{Pages: []Page{{record.Titles.Normalized, strings.TrimSpace(record.Extract), record.Thumbnail.Source, record.ContentUrls.Desktop.Page}}}
		}
	}
}

func processAnalyticsPageviews(body []byte) (list Pagelist) {
	record := AnalyticsPageviews{}
	jsonErr := json.Unmarshal(body, &record)
	if jsonErr != nil || len(record.Items) == 0 {
		fmt.Println(jsonErr)
		if len(record.Detail) != 0 {
			fmt.Println("Error fetching this date. Details:")
			fmt.Println(record.Detail)
		}
		return Pagelist{
			Pages: []PagelistPage{PagelistPage{"Not found.", "", 0, ""}},
		}
	}
	results := record.Items[0].Articles
	collection := []PagelistPage{}
	for _, page := range results {
		articleUrl := fmt.Sprintf(wikiBaseArticlePath, page.Article)

		collection = append(collection, PagelistPage{
			strings.ReplaceAll(page.Article, "_", " "), // Title
			articleUrl,                // URL
			page.Rank,                 // Rank
			strconv.Itoa(page.Views)}) // Pageviews, stringified
	}
	return Pagelist{Pages: collection}

}

// Output the normalized MultiplePages struct with a 'not found' result.
func getNotFound() (pages MultiplePages) {
	return MultiplePages{Pages: []Page{{"Not found.", "", "", ""}}}
}

func getUTCDateToday() (datetime time.Time) {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println(err)
	}
	return time.Now().In(location)
}

func IsDateBeforeUTCToday(requestedDate time.Time) (bigorsmall bool) {
	utcDate := getUTCDateToday()
	// Can't do the direct time comparison (time.Before() time.After())
	// because the actual timestamp doesn't matter, just the year/month/day
	return requestedDate.Year() <= utcDate.Year() && requestedDate.Month() <= utcDate.Month() && requestedDate.Day() <= utcDate.Day()
}
