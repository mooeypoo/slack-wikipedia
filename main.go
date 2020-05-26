package main

import (
	"context"
	"github.com/shomali11/slacker"
	"log"
	"os"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	bot := slacker.NewClient(token)

	definition := &slacker.CommandDefinition{
		Description: "Repeat the text given.",
		Example:     "repeat foo bar baz",
		Handler: func(request slacker.Request, response slacker.ResponseWriter) {
			text := request.StringParam("text", "")
			response.Reply(text)
		},
	}

	bot.Command("echo <text>", definition)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// package main

// import (
// 	"log"
// 	"os"
// 	"strings"

// 	"github.com/sbstjn/hanu"
// )

// func main() {
// 	token := os.Getenv("SLACK_TOKEN")
// 	slack, err := hanu.New(token)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	Version := "0.0.1"

// 	slack.Command("shout <word>", func(conv hanu.ConversationInterface) {
// 		str, _ := conv.String("word")
// 		conv.Reply(strings.ToUpper(str))
// 	})

// 	slack.Command("whisper <word>", func(conv hanu.ConversationInterface) {
// 		str, _ := conv.String("word")
// 		conv.Reply(strings.ToLower(str))
// 	})

// 	slack.Command("version", func(conv hanu.ConversationInterface) {
// 		conv.Reply("Thanks for asking! I'm running `%s`", Version)
// 	})

// 	slack.Listen()
// }
