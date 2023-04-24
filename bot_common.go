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
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"strings"
)

const chatbotHelpText = `
create <repoURL> - Create an OrtRun resource with repoURL
list - List all OrtRun resources
show <name> - Show the nitty gritty of an OrtRun`

var runTableHeaders = []string{"Name", "Scanned Repository", "Analyzer Status", "Scanner Status", "Reporter Status", "Report URL"}

type ortRunMeta struct {
	name           string
	repoUrl        string
	analyzerStatus string
	scannerStatus  string
	reporterStatus string
}

func runListToHtml(unstructuredRuns *unstructured.UnstructuredList) (string, error) {
	runs, err := parseRunList(unstructuredRuns)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	sb.WriteString("<table>")
	sb.WriteString("<tr>")

	for _, header := range runTableHeaders {
		sb.WriteString(fmt.Sprintf("<th>%s</th>", header))
	}

	sb.WriteString("</tr>")

	for _, run := range runs {
		sb.WriteString("<tr>")

		repoUrl := fmt.Sprintf(`<a href="%s">%s</a>`, run.repoUrl, run.repoUrl)

		reportUrl := fmt.Sprintf("https://ortruns.inocybe.io/%s", run.name)
		reportUrl = fmt.Sprintf(`<a href="%s">%s</a>`, reportUrl, reportUrl)

		sb.WriteString(fmt.Sprintf("<td>%s</td>", run.name))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", repoUrl))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", run.analyzerStatus))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", run.scannerStatus))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", run.reporterStatus))
		sb.WriteString(fmt.Sprintf("<td>%s</td>", reportUrl))

		sb.WriteString("</tr>")
	}

	sb.WriteString("</table>")
	return sb.String(), nil
}

func parseRunList(runs *unstructured.UnstructuredList) ([]ortRunMeta, error) {
	parsedRuns := []ortRunMeta{}

	err := runs.EachListItem(func(obj runtime.Object) error {
		item := obj.(*unstructured.Unstructured)

		status, found, err := unstructured.NestedStringMap(item.Object, "status")
		if !found {
			log.Printf("k8s object does not contain field 'status': %v\n", item)
			return nil
		}
		if err != nil {
			log.Printf("fed the wrong thing to unstructured.NestedStringMap(): %v\n", err)
			return nil
		}

		metadata, found, err := unstructured.NestedMap(item.Object, "metadata")
		if !found {
			log.Printf("k8s object does not contain field 'metadata': %v\n", item)
			return nil
		}
		if err != nil {
			log.Printf("fed the wrong thing to unstructured.NestedMap(): %v\n", err)
			return nil
		}

		spec, found, err := unstructured.NestedStringMap(item.Object, "spec")
		if !found {
			log.Printf("k8s object does not contain field 'spec': %v\n", item)
			return nil
		}
		if err != nil {
			log.Printf("fed the wrong thing to unstructured.NestedStringMap(): %v\n", err)
			return nil
		}

		pr := ortRunMeta{}
		pr.name = metadata["name"].(string)
		pr.repoUrl = spec["repoUrl"]
		pr.analyzerStatus = status["analyzer"]
		pr.scannerStatus = status["scanner"]
		pr.reporterStatus = status["reporter"]

		parsedRuns = append(parsedRuns, pr)

		return nil
	})

	return parsedRuns, err
}
