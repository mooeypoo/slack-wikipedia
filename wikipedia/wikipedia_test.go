package wikipedia

import (
	"reflect"
	"testing"
)

func Test_prepTitleForURLQuery(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name          string
		originalTitle string
		expectedTitle string
	}{
		// Test cases
		{"Basic string", "basic", "basic"},
		{"String with spaces", "foo bar", "foo%20bar"},
		{"Keep the initial plus sign", "foo+bar", "foo%2Bbar"},
		{"Keep the initial plus sign with spaces", "foo + bar", "foo%20%2B%20bar"},
		{"Keep capitalization", "FoO bAr", "FoO%20bAr"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTitle := prepTitleForURLQuery(tt.originalTitle); gotTitle != tt.expectedTitle {
				t.Errorf("prepTitleForURLQuery() = %v, want %v", gotTitle, tt.expectedTitle)
			}
		})
	}
}

func Test_processActionAPIResult(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name     string
		body     []byte
		expected []Page
	}{
		{
			"Search with multiple results",
			[]byte(`{"batchcomplete":"","continue":{"gsroffset":3,"continue":"gsroffset||"},"query":{"pages":{"534366":{"pageid":534366,"ns":0,"title":"Title 1","index":1,"extract":"Extract for title 1","thumbnail":{"source":"https://image.for.title/1.png","width":40,"height":50},"pageimage":"https://image.for.title/1/another.jpg","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2020-06-03T02:01:14Z","lastrevid":960453665,"length":356253,"fullurl":"https://xx.wikipedia.org/wiki/Title1","editurl":"https://xx.wikipedia.org/w/index.php?title=Title1&action=edit","canonicalurl":"https://xx.wikipedia.org/wiki/Title1"},"2204744":{"pageid":534366,"ns":0,"title":"Title 2","index":2,"extract":"Extract for title 2","thumbnail":{"source":"https://image.for.title/2.png","width":40,"height":50},"pageimage":"https://image.for.title/2/another.jpg","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2020-06-03T02:01:14Z","lastrevid":960453665,"length":356253,"fullurl":"https://xx.wikipedia.org/wiki/Title2","editurl":"https://xx.wikipedia.org/w/index.php?title=Title2&action=edit","canonicalurl":"https://xx.wikipedia.org/wiki/Title2"},"17775180":{"pageid":534366,"ns":0,"title":"Title 3","index":3,"extract":"Extract for title 3","thumbnail":{"source":"https://image.for.title/3.png","width":40,"height":50},"pageimage":"https://image.for.title/3/another.jpg","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2020-06-03T02:01:14Z","lastrevid":960453665,"length":356253,"fullurl":"https://xx.wikipedia.org/wiki/Title3","editurl":"https://xx.wikipedia.org/w/index.php?title=Title3&action=edit","canonicalurl":"https://xx.wikipedia.org/wiki/Title3"}}}}`),
			[]Page{
				{
					Title:   "Title 1",
					Extract: "Extract for title 1",
					Image:   "https://image.for.title/1.png",
					URL:     "https://xx.wikipedia.org/wiki/Title1",
					Rank:    1,
				},
				{
					Title:   "Title 2",
					Extract: "Extract for title 2",
					Image:   "https://image.for.title/2.png",
					URL:     "https://xx.wikipedia.org/wiki/Title2",
					Rank:    2,
				},
				{
					Title:   "Title 3",
					Extract: "Extract for title 3",
					Image:   "https://image.for.title/3.png",
					URL:     "https://xx.wikipedia.org/wiki/Title3",
					Rank:    3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPage := processActionAPIResult(tt.body); !reflect.DeepEqual(gotPage, tt.expected) {
				t.Errorf("processActionAPIResult() = %v\nExpected:\n %v", gotPage, tt.expected)
			}
		})
	}
}
