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
	"context"
	"fmt"
	"github.com/cip8/autoname"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"strings"
)

const namespace = "ort"

var groupVersionResource = schema.GroupVersionResource{Group: "inocybe.io", Version: "v1", Resource: "ortruns"}

type ortController struct {
	dynClient *dynamic.DynamicClient
	clientset *kubernetes.Clientset
}

func newOrtController() (ortController, error) {
	var oc ortController

	config, err := loadConfig()
	if err != nil {
		return oc, err
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return oc, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return oc, err
	}

	oc.dynClient = dynClient
	oc.clientset = clientset
	return oc, nil
}

func (oc ortController) listRuns() (*unstructured.UnstructuredList, error) {
	return oc.dynClient.
		Resource(groupVersionResource).
		Namespace(namespace).
		List(context.Background(), metav1.ListOptions{})
}

func (oc ortController) getRun(name string) (*unstructured.Unstructured, error) {
	return oc.dynClient.
		Resource(groupVersionResource).
		Namespace(namespace).
		Get(context.Background(), name, metav1.GetOptions{})
}

func (oc ortController) createRun(repoUrl string) (*unstructured.Unstructured, error) {
	name := autoname.Generate("-")

	run := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "inocybe.io/v1",
			"kind":       "OrtRun",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
			"spec": map[string]interface{}{
				"repoUrl": repoUrl,
			},
		},
	}

	created, err := oc.dynClient.
		Resource(groupVersionResource).
		Namespace(namespace).
		Create(context.Background(), run, metav1.CreateOptions{})
	return created, err
}

func (oc ortController) listPods(name, stage string) ([]v1.Pod, error) {
	// TODO only list pods for the OrtRun of the passed in name
	pods, err := oc.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s-%s", stage, name)
	jobPods := []v1.Pod{}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			jobPods = append(jobPods, pod)
		}
	}

	return jobPods, nil
}

func (oc ortController) getLogs(podName string) (string, error) {
	logs, err := oc.clientset.CoreV1().
		Pods(namespace).
		GetLogs(podName, new(v1.PodLogOptions)).
		DoRaw(context.Background())

	if err != nil {
		return "", err
	}
	return string(logs), nil
}

func loadConfig() (*rest.Config, error) {
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))
}
