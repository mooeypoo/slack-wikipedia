package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mooeypoo/slack-wikipedia/wikipedia"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

const resultsLimit = 3

func main() {
	token := os.Getenv("SLACK_TOKEN")
	bot := slacker.NewClient(token)
	fmt.Println("Bot connected.")
	defSummary := &slacker.CommandDefinition{
		Description: "Get the summary of the given page.",
		Example:     "summary san francisco international airport",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			response.Typing()

			text := request.StringParam("text", "")
			result := wikipedia.FetchSummary(text)

			attachments := getFullReplyAttachments(text, fmt.Sprintf("Here's the page for \"*%s*\" on Wikipedia:", text), result)
			response.Reply(text, slacker.WithBlocks(attachments))
		},
	}

	defRelated := &slacker.CommandDefinition{
		Description: "Find articles that are related to your search.",
		Example:     "related Barack Obama",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			response.Typing()

			text := request.StringParam("text", "")
			results := wikipedia.FetchRelated(text)

			attachments := getFullReplyAttachments(text, fmt.Sprintf("Here's are some Wikipedia articles related to \"*%s*\":", text), results)
			response.Reply(text, slacker.WithBlocks(attachments))
		},
	}

	defSearch := &slacker.CommandDefinition{
		Description: "Search for Wikipedia articles.",
		Example:     "search summer vacation",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			response.Typing()

			text := request.StringParam("text", "")
			results := wikipedia.FetchSearch(text)

			attachments := getFullReplyAttachments(text, fmt.Sprintf("Here's what I found for \"*%s*\" on Wikipedia:", text), results)
			response.Reply(text, slacker.WithBlocks(attachments))
		},
	}

	defTopviews := &slacker.CommandDefinition{
		Description: "See top viewed articles for the given date. Provide no date to see today's results.",
		Example:     "top March 1 2020",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			response.Typing()

			text := request.StringParam("text", "")
			actualRequestedTime := wikipedia.ParseTimeString(text)

			// Build output
			attachments := []slack.Block{}

			if wikipedia.IsDateBeforeUTCToday(actualRequestedTime) {
				// We are asking for a date that is still "tomorrow" for UTC
				// Change that to the previous day
				newRequestedTime := actualRequestedTime.AddDate(0, 0, -1)

				// Let the user know, but only if the user actually requested a date and not an empty string
				if len(strings.TrimSpace(text)) != 0 {
					humanReadableOrig := fmt.Sprintf("%s %02d %d", actualRequestedTime.Month(), actualRequestedTime.Day(), actualRequestedTime.Year())
					humanReadableNew := fmt.Sprintf("%s %02d %d", newRequestedTime.Month(), newRequestedTime.Day(), newRequestedTime.Year())

					switchDateText := slack.NewTextBlockObject("mrkdwn",
						fmt.Sprintf("I don't have information yet for the top views on *%s*. Let's see if I can find any results for *%s* instead.", humanReadableOrig, humanReadableNew),
						false, false)

					switchDateSection := slack.NewSectionBlock(switchDateText, nil, nil)
					attachments = append(attachments, switchDateSection)
				}
				actualRequestedTime = newRequestedTime
			}

			formattedRequestedTime := fmt.Sprintf("%s %02d %d", actualRequestedTime.Month(), actualRequestedTime.Day(), actualRequestedTime.Year())

			results, _ := wikipedia.FetchTopPageviews(formattedRequestedTime)
			fmt.Printf("Requested 'top' with parameter \"%s\" parsed into date \"%s\"\n", text, formattedRequestedTime)

			if len(results) == 0 || results[0].Title == "" || results[0].Title == "Not found." {
				notFoundText := slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("Oops, I couldn't find the top viewed articles for the date *\"%s\"*. :face_with_rolling_eyes: :grimacing:", formattedRequestedTime),
					false, false)
				fmt.Println("Request for top views not found.")
				headerSection := slack.NewSectionBlock(notFoundText, nil, nil)
				attachments = append(attachments, headerSection)
			} else {
				fmt.Println("Request for top views found. Response being built.")
				// Remove 'Main_Page' and 'Special:Search' from results
				// And cut at 10 items
				header := slack.NewSectionBlock(slack.NewTextBlockObject(
					"mrkdwn",
					fmt.Sprintf("Top viewed pages for *%s* _(Not including the Main and Search pages)_", formattedRequestedTime),
					false, false),
					nil, nil)
				attachments = append(attachments, header)

				for _, page := range results {
					if page.Title != "Main Page" && page.Title != "Special:Search" && len(attachments) < 10 {
						section := slack.NewSectionBlock(slack.NewTextBlockObject(
							"mrkdwn",
							fmt.Sprintf("*%d most viewed:* <%s|%s> (%s page views)", page.Rank, page.URL, page.Title, page.Info),
							false, false),
							nil, nil)
						attachments = append(attachments, section)
					}
				}
			}
			fmt.Printf("Sending response to Slack with %d attachments\n", len(attachments))
			response.Reply(formattedRequestedTime, slacker.WithBlocks(attachments))
		},
	}

	bot.Command("summary <text>", defSummary)
	bot.Command("related <text>", defRelated)
	bot.Command("search <text>", defSearch)
	bot.Command("top <text>", defTopviews)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// Build the reply attachments for the commands, and answer properly
// when a search text query was not found.
func getFullReplyAttachments(searchText string, headerText string, results []wikipedia.Page) (att []slack.Block) {
	attachments := []slack.Block{}
	if len(strings.TrimSpace(searchText)) == 0 {
		notFoundText := slack.NewTextBlockObject("mrkdwn",
			"Give me something to look up...?",
			false, false)
		attachments = append(attachments, notFoundText)
		return attachments
	}
	if len(results) == 0 || results[0].Title == "" || results[0].Title == "Not found." {
		notFoundText := slack.NewTextBlockObject("mrkdwn",
			fmt.Sprintf("I couldn't find anything related to \"*%s*\" :face_with_rolling_eyes: :grimacing:", searchText),
			false, false)
		attachments = append(attachments, notFoundText)
		return attachments
	}

	// Create the formatted response
	header := getResultListHeader(headerText)
	list := buildResultListAttachments(results)
	attachments = append(attachments, header...)
	attachments = append(attachments, list...)

	return attachments
}

// Add a header for the search results with a given text
// attatchment parameter is the existing array of blocks from the result list
func getResultListHeader(headerStringText string) (att []slack.Block) {
	headerText := slack.NewTextBlockObject("mrkdwn",
		headerStringText,
		false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)
	divSection := slack.NewDividerBlock()

	// Prepend to attachments
	attachments := []slack.Block{}
	attachments = append(attachments, headerSection, divSection)

	return attachments
}

// Build a slack block list from the results from the API
func buildResultListAttachments(results []wikipedia.Page) (att []slack.Block) {

	infoTextPrintf := ""
	if len(results) > 1 {
		// For multiple results, limit the extract
		infoTextPrintf = "*<%s|%s>*\n%.200s[...]"
	} else {
		// For one result, don't limit the extract
		infoTextPrintf = "*<%s|%s>*\n%s"
	}

	// Create the formatted response
	attachments := []slack.Block{}

	// Build the list
	for index, page := range results {
		if index >= resultsLimit {
			break
		}
		imageURL := ""
		if page.Image != "" {
			imageURL = page.Image
		} else {
			imageURL = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f1/No-image.svg/200px-No-image.svg.png"
		}
		attachments = append(attachments, slack.NewSectionBlock(
			// Item info
			slack.NewTextBlockObject(
				"mrkdwn",
				fmt.Sprintf(infoTextPrintf, page.URL, page.Title, page.Extract),
				false, false),
			nil,
			slack.NewAccessory(slack.NewImageBlockElement(imageURL, page.Title))))
	}

	return attachments
}
