package main

import (
	"log"
	"os"
	"strings"

	"github.com/sbstjn/hanu"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")

	// slack, err := hanu.New("xoxb-1143242370325-1144948666467-D7affrVWkgWE2o3fghcPcYnq")
	slack, err := hanu.New(token)

	if err != nil {
		log.Fatal(err)
	}

	Version := "0.0.1"

	slack.Command("shout <word>", func(conv hanu.ConversationInterface) {
		str, _ := conv.String("word")
		conv.Reply(strings.ToUpper(str))
	})

	slack.Command("whisper <word>", func(conv hanu.ConversationInterface) {
		str, _ := conv.String("word")
		conv.Reply(strings.ToLower(str))
	})

	slack.Command("version", func(conv hanu.ConversationInterface) {
		conv.Reply("Thanks for asking! I'm running `%s`", Version)
	})

	slack.Listen()
}
