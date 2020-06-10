[![Build Status](https://travis-ci.com/mooeypoo/slack-wikipedia.svg?branch=master)](https://travis-ci.com/mooeypoo/slack-wikipedia)

Wikipedia Slackbot
==================

A Slack bot that fetches information from Wikipedia APIs.
This is a work in progress, please beware running this in a public Slack workspace. More info will be added to the readme soon.

== Running the bot
To run the bot locally:

1. Clone the repo
2. Get a bot token for your Slack token
3. Add a local variable `SLACK_TOKEN` with the value of the token you created
2. Run `go run main.go`

= Bot commands
To see the list of available commands, mention the bot with `help`. Example: `@wikibot help`. 

To respond to commands, the bot needs to either be in a channel it was directly invited into, or the command needs to be given in a private message to the bot user.

= Credits and license
Created by Moriel Schottlender (mooeypoo) under GPLv3 license.

Please report bugs and feature requests in the issues!


