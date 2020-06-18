package wikipedia

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/araddon/dateparse"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TODO: Maybe make this configurable? Allow for other languages?
var wikiBaseArticlePath = "https://%s.wikipedia.org/wiki/%s"
var wikiRESTEndpoint = "https://%s.wikipedia.org/api/rest_v1/"
var wikiRESTsummary = "page/summary/%s?redirect=true"
var wikiRESTrelated = "page/related/%s"
var wikiActionAPIendpoint = "https://%s.wikipedia.org/w/api.php"
var wikiAnalyticsPageviewsEndpoint = "https://wikimedia.org/api/rest_v1/metrics/pageviews/top/%s.wikipedia/all-access/%d/%02d/%02d" // "2020/06/02"

// Page is a normalized structure for representing page data
type Page struct {
	Title   string
	Extract string
	Image   string
	URL     string
	Rank    int
}

// PagelistPage represent normalized structure for an information for a page in a list
type PagelistPage struct {
	Title string
	URL   string
	Rank  int
	Info  string
}

// FetchSummary fetches the summary of a specific Wikipedia page given by its title
func FetchSummary(title string) (resp []Page, lang string, actualTitle string) {
	lang, strippedTitle := ParseLanguageFromText(title)
	safeTitle := prepTitleForURLQuery(strippedTitle)

	url := fmt.Sprintf(wikiRESTEndpoint+wikiRESTsummary, lang, safeTitle)
	toLog("FetchSummary", url)

	body, readErr := fetchFromAPI(url)
	if readErr != nil {
		return getNotFound(), lang, strippedTitle
	}

	return processRESTApiResult(body, false), lang, strippedTitle
}

// FetchRelated fetches the related pages for the given term
func FetchRelated(term string) (resp []Page, lang string, actualTerm string) {
	lang, strippedTerm := ParseLanguageFromText(term)
	safeTitle := prepTitleForURLQuery(strippedTerm)

	url := fmt.Sprintf(wikiRESTEndpoint+wikiRESTrelated, lang, safeTitle)
	toLog("FetchRelated", "URL: "+url)

	body, readErr := fetchFromAPI(url)
	if readErr != nil {
		return getNotFound(), lang, strippedTerm
	}

	return processRESTApiResult(body, true), lang, strippedTerm
}

// FetchSearch fetches search results from Wikipedia given the search string
func FetchSearch(searchString string) (resp []Page, lang string, actualSearchString string) {
	lang, strippedTerm := ParseLanguageFromText(searchString)

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
	// params.Add("gsrwhat", "text")
	params.Add("gsrwhat", "nearmatch")
	params.Add("gsrsearch", strippedTerm)

	url := fmt.Sprintf(wikiActionAPIendpoint, lang) + "?" + params.Encode()
	toLog("FetchSearch", "URL: "+url)

	body, readErr := fetchFromAPI(url)
	if readErr != nil {
		return getNotFound(), lang, strippedTerm
	}

	return processActionAPIResult(body), lang, strippedTerm
}

// FetchTopPageviews fetches the top articles by pageview for a given date.
// Lang parameter will dictate the Wikipedia that will be searched. If given
// empty string, will fall back on "en"
func FetchTopPageviews(datestring string, lang string) (resp []PagelistPage) {
	t := ParseTimeString(datestring)

	if len(lang) == 0 {
		lang = "en"
	}
	// Build the url
	url := fmt.Sprintf(wikiAnalyticsPageviewsEndpoint, lang, t.Year(), int(t.Month()), t.Day())

	toLog("FetchTopPageviews", "URL: "+url)

	body, readErr := fetchFromAPI(url)
	if readErr != nil {
		return []PagelistPage{{"Not found.", "", 0, ""}}
	}

	return processAnalyticsPageviews(body, lang)
}

// FetchGetGeneralTerm is a general method of fetching a term from Wikipedia,
// displaying related terms, and showing a general search result if a page
// isn't found.
// This is mostly meant to make sure that a user expects -- and gets -- sensible
// results from Wikipedia regardless of the "quirks" of the API mechanisms like
// expecting exact character chase or exact match for titles, and also suggesting
// related information rather than leaving the user with no result at all.
//
// = Fallback mechanism
// The process performs the following with the given term
// - Always: Fetch the summary of the <term>
//   - If summary found:
//     - Fetch related articles for the <term>
//   - If summary not found:
//     - Fetch search results for the <term>
//     - If first result (by relevance) is equal (regardless of case) to the
//       requested term, continue as if summary was found.
//     - Otherwise, return search results
//
// = Return value
// The method returns a list of results, and a list of 'sub' results (related pages)
// so the consumer can display those differently if they wish.
func FetchGetGeneralTerm(term string) (results []Page, related []Page, lang string, actualTitle string) {
	relatedPages := []Page{}
	summaryPages, lang, actualTitle := FetchSummary(term)
	toLog("FetchGetGeneralTerm term", term)
	if summaryPages[0].Title != "Not found." {
		toLog("FetchGetGeneralTerm summary found", summaryPages[0].Title)
		// Page found. Fetch related
		relatedPages, _, _ = FetchRelated(summaryPages[0].Title)
		return summaryPages, relatedPages, lang, actualTitle
	}
	toLog("FetchGetGeneralTerm summary not found for title", actualTitle)

	// Summary wasn't found. Do a search
	searchPages, _, _ := FetchSearch(term)
	if searchPages[0].Title != "Not found." {
		toLog("FetchGetGeneralTerm search found with "+strconv.Itoa(len(searchPages))+" results", searchPages[0].Title)
		if len(searchPages) == 1 || strings.ToLower(searchPages[0].Title) == strings.ToLower(actualTitle) {
			// This is the page we're looking for. Fetch related to the actual title
			relatedPages, _, _ = FetchRelated(searchPages[0].Title)

			// Only return the first page
			searchPages = append([]Page{}, searchPages[:1]...)
			toLog("FetchGetGeneralTerm returning first page of search results", searchPages[0].Title)
			return searchPages, relatedPages, lang, actualTitle
		} else {
			// Return the search results
			toLog("FetchGetGeneralTerm", "Returning full search results")
			return searchPages, []Page{}, lang, actualTitle
		}
	}

	// Search results not found. Return 'not found'
	toLog("FetchGetGeneralTerm not found: ", term)
	return summaryPages, searchPages, lang, actualTitle
}

// ParseTimeString normalizes and then parses the given string into a time object
func ParseTimeString(datestring string) (parsed time.Time) {
	datestring = strings.TrimSpace(datestring)
	fmt.Printf("Parsing time string \"%s\"", datestring)
	t := time.Now()
	if len(datestring) != 0 {
		parsedTime, err := dateparse.ParseAny(datestring)
		if err != nil {
			t = time.Now()
			toLog("ParseTimeString", "Failed to parse given date string: "+datestring)
			// fmt.Println(err)
		} else {
			t = parsedTime
		}
	}
	return t
}

// ParseLanguageFromText looks for the lang=xx expression and outputs
// the language, or defaults to 'en' if language wasn't found.
func ParseLanguageFromText(text string) (lang string, remainingText string) {
	r, _ := regexp.Compile("lang=([[:alpha:]_-]+)")
	match := r.FindStringSubmatch(text)

	if len(match) > 0 {
		// Remove that from the string
		newText := strings.TrimSpace(r.ReplaceAllString(text, ""))
		return match[1], newText
	}
	return "en", strings.TrimSpace(text)
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
func fetchFromAPI(link string) (body []byte, err error) {
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
// data structure through Page structure
func processActionAPIResult(body []byte) (page []Page) {
	record := ActionAPIGeneratorResponse{}
	jsonErr := json.Unmarshal(body, &record)
	if jsonErr != nil || len(record.Query.Pages) == 0 {
		return getNotFound()
	}

	collection := []Page{}
	for _, page := range record.Query.Pages {
		collection = append(collection, Page{
			page.Title,
			strings.TrimSpace(page.Extract),
			page.Thumbnail.Source,
			page.Canonicalurl,
			page.Index})
	}
	sort.SliceStable(collection, func(i, j int) bool {
		return collection[i].Rank < collection[j].Rank
	})
	return collection
}

// Get result from the Wikipedia RESTBASE API and output a normalized
// data structure through multiple Page output
// isMultple parameter should be set to true if the request is expected
// to return a JSON structure that holds multiple results. False otherwise.
func processRESTApiResult(body []byte, isMultiple bool) (page []Page) {
	if isMultiple {
		record := MultiplePageResponseREST{}
		jsonErr := json.Unmarshal(body, &record)
		if jsonErr != nil || len(record.Pages) == 0 {
			return getNotFound()
		}

		collection := []Page{}
		for _, page := range record.Pages {
			collection = append(collection, Page{
				page.Titles.Normalized,
				strings.TrimSpace(page.Extract),
				page.Thumbnail.Source,
				page.ContentUrls.Desktop.Page,
				0})
		}
		return collection
	}

	// Single result
	record := PageResponseREST{}
	jsonErr := json.Unmarshal(body, &record)
	if jsonErr != nil || record.Title == "Not found." {
		return getNotFound()
	}
	return []Page{{record.Titles.Normalized, strings.TrimSpace(record.Extract), record.Thumbnail.Source, record.ContentUrls.Desktop.Page, 0}}
}

// Process the result from the Wikipedia analytics Pageview API endpoint
// and return a list representing the pages with their pageview and rank
func processAnalyticsPageviews(body []byte, lang string) (list []PagelistPage) {
	record := AnalyticsPageviews{}
	jsonErr := json.Unmarshal(body, &record)
	if jsonErr != nil || len(record.Items) == 0 {
		if len(record.Detail) != 0 {
			toLog("processAnalyticsPageviews", "Error fetching. Details: "+record.Detail)
		}
		return []PagelistPage{{"Not found.", "", 0, ""}}
	}
	results := record.Items[0].Articles
	collection := []PagelistPage{}
	for _, page := range results {
		articleURL := fmt.Sprintf(wikiBaseArticlePath, lang, url.QueryEscape(page.Article))

		collection = append(collection, PagelistPage{
			strings.ReplaceAll(page.Article, "_", " "), // Title
			articleURL,                // URL
			page.Rank,                 // Rank
			strconv.Itoa(page.Views)}) // Pageviews, stringified
	}
	return collection
}

// IsDateBeforeUTCToday checks whether the given date is before the official "today" date of UTC.
//
// This is meant to see if there needs to be a conversion (going 'back' a day)
// when the user from a timezone that is ahead of UTC requests data for a certain
// date. For example, a user in San Francisco asking for results for June 5th,
// The UTC date may still be June 4th, which means the results from the remote
// API request (specifically analytics clusters, but others may as well) be
// unavailable. If that is the case, this gives the consumer a chance to change
// the date to a day before or alert the user that they should change their
// requested date themselves.
func IsDateBeforeUTCToday(requestedDate time.Time) (isBefore bool) {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return true
	}
	utcDate := time.Now().In(location)
	// Can't do the direct time comparison (time.Before() time.After())
	// because the actual timestamp doesn't matter, just the year/month/day
	return requestedDate.Year() <= utcDate.Year() && requestedDate.Month() <= utcDate.Month() && requestedDate.Day() < utcDate.Day()
}

// Output the normalized structure with a 'not found' result.
func getNotFound() (pages []Page) {
	return []Page{{"Not found.", "", "", "", 0}}
}

// Output to a log, including timestamps and context
// For the moment, print this out.
func toLog(context string, str string) {
	now := time.Now()

	fmt.Println("[" + now.Format(time.RFC822) + "] " + context + ": " + str)
}
