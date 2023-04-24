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
	"log"
	"net/http"
	"os"
)

var (
	matrixServer      = os.Getenv("MATRIX_SERVER")
	matrixUser        = os.Getenv("MATRIX_USER")
	matrixAccessToken = os.Getenv("MATRIX_ACCESS_TOKEN")
)

func main() {
	if matrixServer != "" {
		bot, err := newMatrixBot(matrixServer, matrixUser, matrixAccessToken)
		if err != nil {
			log.Fatal(err)
		}

		bot.run()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/runs", handleRuns)
	mux.HandleFunc("/runs/", handleGetRun)
	mux.HandleFunc("/logs/", handleLogs)

	log.Print("Starting server on :4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}
