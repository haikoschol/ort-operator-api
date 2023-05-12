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
	"log"
	"net/http"
	"strings"
)

func handleRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleCorsRequest(w, "GET,POST")
		return
	}

	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed, "GET,POST")
		return
	}

	oc, err := newOrtController()
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET,POST")
		return
	}

	if r.Method == http.MethodPost {
		handleCreateRun(w, r, oc)
		return
	}

	if r.Method == http.MethodGet {
		handleListRuns(w, oc)
		return
	}
}

func handleCreateRun(w http.ResponseWriter, r *http.Request, oc ortController) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)

	payload := struct {
		RepoUrl string `json:"repoUrl"`
	}{}

	if err := decoder.Decode(&payload); err != nil {
		writeStatus(w, http.StatusBadRequest, "GET,POST")
		return
	}

	created, err := oc.createRun(payload.RepoUrl)
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET,POST")
		return
	}

	ortRun, err := unstructuredToOrtRun(created, true)
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET,POST")
		return
	}

	writeCorsHeaders(w, "GET,POST")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(ortRun); err != nil {
		log.Println(err)
	}
}

func handleListRuns(w http.ResponseWriter, oc ortController) {
	runs, err := oc.listRuns()
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	ortRuns, err := unstructuredListToOrtRunList(runs)
	if err != nil {
		log.Println(err)
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	writeCorsHeaders(w, "GET")
	w.Header().Add("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(ortRuns); err != nil {
		log.Println(err)
	}
}

func handleGetRun(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleCorsRequest(w, "GET")
		return
	}

	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed, "GET")
		return
	}

	oc, err := newOrtController()
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	name, found := strings.CutPrefix(r.URL.Path, "/runs/")
	if !found || name == "" {
		writeStatus(w, http.StatusBadRequest, "GET")
		return
	}

	run, err := oc.getRun(name)
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	ortRun, err := unstructuredToOrtRun(run, true)
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	writeCorsHeaders(w, "GET")
	w.Header().Add("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(ortRun); err != nil {
		log.Println(err)
	}
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleCorsRequest(w, "GET")
		return
	}

	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed, "GET")
		return
	}

	path, found := strings.CutPrefix(r.URL.Path, "/logs/")
	if !found {
		writeStatus(w, http.StatusBadRequest, "GET")
		return
	}

	name, stage, found := strings.Cut(path, "/")
	if !found {
		writeStatus(w, http.StatusBadRequest, "GET")
		return
	}

	oc, err := newOrtController()
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	pods, err := oc.listPods(name, stage)
	if err != nil {
		writeStatus(w, http.StatusInternalServerError, "GET")
		return
	}

	var podLogs []PodLogs

	for _, pod := range pods {
		logs, err := oc.getLogs(pod.Name)
		if err != nil {
			writeStatus(w, http.StatusInternalServerError, "GET")
			return
		}

		podLogs = append(podLogs, PodLogs{PodName: pod.Name, PodLogs: logs})
	}

	ortLogs := OrtLogs{
		RunName:  name,
		RunStage: stage,
		PodLogs:  podLogs,
	}

	writeCorsHeaders(w, "GET")
	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(ortLogs); err != nil {
		log.Println(err)
	}
}

func writeStatus(w http.ResponseWriter, status int, allowedMethods string) {
	writeCorsHeaders(w, allowedMethods)
	w.WriteHeader(status)
}

func handleCorsRequest(w http.ResponseWriter, allowedMethods string) {
	writeStatus(w, http.StatusNoContent, allowedMethods)
}

func writeCorsHeaders(w http.ResponseWriter, allowedMethods string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", allowedMethods)
}
