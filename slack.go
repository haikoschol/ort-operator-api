// Copyright (C) 2023 Haiko Schol
// SPDX-License-Identifier: GPL-3.0-or-later

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
)

var (
	slackToken = os.Getenv("SLACK_TOKEN")
	channelId  = os.Getenv("SLACK_CHANNEL_ID")
)

type slackBot struct {
	oc        ortController
	smHandler *socketmode.SocketmodeHandler
}

func newSlackBot() (slackBot, error) {
	oc, err := newOrtController()
	if err != nil {
		return slackBot{}, err
	}

	client := slack.New(slackToken)

	attachment := slack.Attachment{
		Text: "test",
	}

	_, _, err = client.PostMessage(channelId, slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Printf("failed to post a message to Slack: %v\n", err)
	}

	socketmodeClient := socketmode.New(client)
	smHandler := socketmode.NewSocketmodeHandler(socketmodeClient)

	return slackBot{
		oc,
		smHandler,
	}, nil
}

func (sb slackBot) run() {
	sb.smHandler.HandleEvents(slackevents.AppMention, sb.handleMention)
	go func() {
		if err := sb.smHandler.RunEventLoop(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (sb slackBot) handleMention(event *socketmode.Event, client *socketmode.Client) {
	attachment := slack.Attachment{
		Text: chatbotHelpText,
	}

	_, _, err := client.PostMessage(channelId, slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Printf("failed to post a message to Slack: %v\n", err)
	}
}
