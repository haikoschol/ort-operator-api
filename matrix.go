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
	"encoding/json"
	"fmt"
	"github.com/matrix-org/gomatrix"
	"log"
	"strings"
)

type matrixBot struct {
	oc     ortController
	client *gomatrix.Client
}

func newMatrixBot(server, user, accessToken string) (matrixBot, error) {
	oc, err := newOrtController()
	if err != nil {
		return matrixBot{}, err
	}

	client, err := gomatrix.NewClient(server, user, accessToken)
	if err != nil {
		return matrixBot{}, err
	}

	return matrixBot{
		oc,
		client,
	}, nil
}

func (mb matrixBot) run() {
	syncer := mb.client.Syncer.(*gomatrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", mb.handleMessage)

	// TODO teardown
	go func() {
		for {
			if err := mb.client.Sync(); err != nil {
				log.Printf("failed to sync state with matrix server %v: %v\n", mb.client.HomeserverURL, err)
			}
		}
	}()
}

func (mb matrixBot) handleMessage(ev *gomatrix.Event) {
	body, ok := ev.Body()
	if !ok {
		return
	}

	suffix := fmt.Sprintf(":%s", mb.client.HomeserverURL.Host)
	shortUid, _ := strings.CutSuffix(mb.client.UserID, suffix)
	body, found := strings.CutPrefix(body, shortUid)
	if !found {
		return
	}

	body = strings.TrimSpace(body)
	cmd, args, _ := strings.Cut(body, " ")
	cmd = strings.TrimSpace(cmd)
	args = strings.TrimSpace(args)

	mb.handleCommand(ev, cmd, args)
}

func (mb matrixBot) handleCommand(ev *gomatrix.Event, command, arguments string) {
	switch command {
	case "help":
		mb.sendCommandResponse(ev, chatbotHelpText)
	case "create":
		mb.handleCreateCommand(ev, arguments)
	case "list":
		mb.handleListCommand(ev)
	case "show":
		mb.handleShowCommand(ev, arguments)
	default:
		message := fmt.Sprintf("unknown command '%s'. Use 'help' to list all available commands", command)
		mb.sendCommandResponse(ev, message)
	}
}

func (mb matrixBot) handleCreateCommand(ev *gomatrix.Event, repoUrl string) {
	run, err := mb.oc.createRun(repoUrl)
	if err != nil {
		mb.sendCommandResponse(ev, err.Error())
		return
	}

	data, err := json.MarshalIndent(run, "", "    ")
	if err != nil {
		mb.sendCommandResponse(ev, err.Error())
		return
	}

	mb.sendCommandResponse(ev, string(data))
}

func (mb matrixBot) handleListCommand(ev *gomatrix.Event) {
	runs, err := mb.oc.listRuns()
	if err != nil {
		mb.sendCommandResponse(ev, err.Error())
		return
	}

	html, err := runListToHtml(runs)
	if err != nil {
		mb.sendCommandResponse(ev, err.Error())
		return
	}

	html = fmt.Sprintf("%s %s", ev.Sender, html)

	if _, err = mb.client.SendFormattedText(ev.RoomID, "", html); err != nil {
		log.Println("failed to send run list as HTML to Matrix: %v\n", err)
	}
}

func (mb matrixBot) handleShowCommand(ev *gomatrix.Event, name string) {
	run, err := mb.oc.getRun(name)
	if err != nil {
		mb.sendCommandResponse(ev, err.Error())
		return
	}

	data, err := json.MarshalIndent(run, "", "    ")
	if err != nil {
		mb.sendCommandResponse(ev, err.Error())
		return
	}

	mb.sendCommandResponse(ev, string(data))
}

// sendCommandResponse sends the message to the sender of the command and only logs locally in case of error
func (mb matrixBot) sendCommandResponse(ev *gomatrix.Event, message string) {
	message = fmt.Sprintf("%s %s", ev.Sender, message)

	_, err := mb.client.SendText(ev.RoomID, message)
	if err != nil {
		log.Printf(
			"failed to send command response to matrix server %v. error: '%v' message: '%s'\n",
			mb.client.HomeserverURL,
			err,
			message,
		)
	}
}
