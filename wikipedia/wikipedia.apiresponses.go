package wikipedia

import (
	"time"
)

// API Response structs
type ActionAPIGeneratorResponse struct {
	Batchcomplete string `json:"batchcomplete"`
	Continue      struct {
		Gsroffset int    `json:"gsroffset"`
		Continue  string `json:"continue"`
	} `json:"continue"`
	Query struct {
		Pages map[string]ActionAPIBaseResponsePageInfo `json:"pages"`
	} `json:"query"`
}

type ActionAPIBaseResponsePageInfo struct {
	Pageid    int    `json:"pageid"`
	Ns        int    `json:"ns"`
	Title     string `json:"title"`
	Index     int    `json:"index"`
	Extract   string `json:"extract"`
	Thumbnail struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail"`
	Pageimage            string    `json:"pageimage"`
	Contentmodel         string    `json:"contentmodel"`
	Pagelanguage         string    `json:"pagelanguage"`
	Pagelanguagehtmlcode string    `json:"pagelanguagehtmlcode"`
	Pagelanguagedir      string    `json:"pagelanguagedir"`
	Touched              time.Time `json:"touched"`
	Lastrevid            int       `json:"lastrevid"`
	Length               int       `json:"length"`
	Fullurl              string    `json:"fullurl"`
	Editurl              string    `json:"editurl"`
	Canonicalurl         string    `json:"canonicalurl"`
}

type MultiplePageResponseREST struct {
	Pages []PageResponseREST `json:"pages"`
}

type PageResponseREST struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	Displaytitle string `json:"displaytitle"`
	Namespace    struct {
		ID   int    `json:"id"`
		Text string `json:"text"`
	} `json:"namespace"`
	WikibaseItem string `json:"wikibase_item"`
	Titles       struct {
		Canonical  string `json:"canonical"`
		Normalized string `json:"normalized"`
		Display    string `json:"display"`
	} `json:"titles"`
	Pageid    int `json:"pageid"`
	Thumbnail struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail"`
	Originalimage struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"originalimage"`
	Lang              string `json:"lang"`
	Dir               string `json:"dir"`
	Revision          string `json:"revision"`
	Tid               string `json:"tid"`
	Timestamp         string `json:"timestamp"`
	Description       string `json:"description"`
	DescriptionSource string `json:"description_source"`
	Coordinates       struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
	ContentUrls struct {
		Desktop struct {
			Page      string `json:"page"`
			Revisions string `json:"revisions"`
			Edit      string `json:"edit"`
			Talk      string `json:"talk"`
		} `json:"desktop"`
		Mobile struct {
			Page      string `json:"page"`
			Revisions string `json:"revisions"`
			Edit      string `json:"edit"`
			Talk      string `json:"talk"`
		} `json:"mobile"`
	} `json:"content_urls"`
	APIUrls struct {
		Summary      string `json:"summary"`
		Metadata     string `json:"metadata"`
		References   string `json:"references"`
		Media        string `json:"media"`
		EditHTML     string `json:"edit_html"`
		TalkPageHTML string `json:"talk_page_html"`
	} `json:"api_urls"`
	Extract     string `json:"extract"`
	ExtractHTML string `json:"extract_html"`
}

type AnalyticsPageviews struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Items  []struct {
		Project  string `json:"project"`
		Access   string `json:"access"`
		Year     string `json:"year"`
		Month    string `json:"month"`
		Day      string `json:"day"`
		Articles []struct {
			Article string `json:"article"`
			Views   int    `json:"views"`
			Rank    int    `json:"rank"`
		} `json:"articles"`
	} `json:"items"`
}
