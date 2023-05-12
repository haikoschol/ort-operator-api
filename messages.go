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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type StageStatus int

const (
	Pending StageStatus = iota
	Running
	Succeeded
	Failed
	Aborted
)

type RunStatus struct {
	Analyzer StageStatus `json:"analyzer"`
	Scanner  StageStatus `json:"scanner"`
	Reporter StageStatus `json:"reporter"`
}

type OrtRun struct {
	Name               string    `json:"name"`
	RepoUrl            string    `json:"repoUrl"`
	Status             RunStatus `json:"status"`
	KubernetesResource string    `json:"kubernetesResource,omitempty"`
}

type OrtRunList struct {
	Runs []OrtRun `json:"runs"`
}

// FIXME bad naming: "PodLogs" used twice
type PodLogs struct {
	PodName string `json:"podName"`
	PodLogs string `json:"podLogs"`
}

type OrtLogs struct {
	RunName  string    `json:"runName"`
	RunStage string    `json:"runStage"`
	PodLogs  []PodLogs `json:"podLogs"`
}

func unstructuredToOrtRun(resource *unstructured.Unstructured, withYaml bool) (OrtRun, error) {
	run := OrtRun{
		Name:    resource.Object["metadata"].(map[string]interface{})["name"].(string),
		RepoUrl: resource.Object["spec"].(map[string]interface{})["repoUrl"].(string),
		Status:  runStatusFromUnstructured(resource),
	}

	if withYaml {
		objYaml, err := yaml.Marshal(resource.Object)
		if err != nil {
			return run, err
		}

		run.KubernetesResource = string(objYaml)
	}

	return run, nil
}

func unstructuredListToOrtRunList(list *unstructured.UnstructuredList) (OrtRunList, error) {
	runList := OrtRunList{}
	var runs []OrtRun

	for _, item := range list.Items {
		run, err := unstructuredToOrtRun(&item, false)
		if err != nil {
			return runList, err
		}

		runs = append(runs, run)
	}

	runList.Runs = runs
	return runList, nil
}

func (s StageStatus) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Running:
		return "Running"
	case Succeeded:
		return "Succeeded"
	case Failed:
		return "Failed"
	case Aborted:
		return "Aborted"
	}
	return "Unknown"
}

func runStatusFromUnstructured(resource *unstructured.Unstructured) RunStatus {
	result := RunStatus{Pending, Pending, Pending}
	statuses, found := resource.Object["status"].(map[string]interface{})
	if !found {
		return result
	}

	s, found := statuses["analyzer"].(string)
	if found {
		result.Analyzer = stageStatusFromString(s)
	}

	s, found = statuses["scanner"].(string)
	if found {
		result.Scanner = stageStatusFromString(s)
	}

	s, found = statuses["reporter"].(string)
	if found {
		result.Reporter = stageStatusFromString(s)
	}

	return result
}

func (s StageStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func stageStatusFromString(s string) StageStatus {
	switch s {
	case "Running":
		return Running
	case "Succeeded":
		return Succeeded
	case "Failed":
		return Failed
	case "Aborted":
		return Aborted
	default:
		return Pending
	}
}
