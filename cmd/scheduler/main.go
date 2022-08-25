/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	"k8s.io/component-base/cli"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/freckie/shmsched-plugin/pkg/capacityscheduling"
	"github.com/freckie/shmsched-plugin/pkg/coscheduling"
	"github.com/freckie/shmsched-plugin/pkg/noderesources"
	"github.com/freckie/shmsched-plugin/pkg/noderesourcetopology"
	"github.com/freckie/shmsched-plugin/pkg/podstate"
	"github.com/freckie/shmsched-plugin/pkg/preemptiontoleration"
	"github.com/freckie/shmsched-plugin/pkg/qos"
	"github.com/freckie/shmsched-plugin/pkg/shmscoring"
	"github.com/freckie/shmsched-plugin/pkg/trimaran/loadvariationriskbalancing"
	"github.com/freckie/shmsched-plugin/pkg/trimaran/targetloadpacking"

	// Ensure scheme package is initialized.
	_ "github.com/freckie/shmsched-plugin/apis/config/scheme"
)

func main() {
	// Register custom plugins to the scheduler framework.
	// Later they can consist of scheduler profile(s) and hence
	// used by various kinds of workloads.
	command := app.NewSchedulerCommand(
		app.WithPlugin(capacityscheduling.Name, capacityscheduling.New),
		app.WithPlugin(coscheduling.Name, coscheduling.New),
		app.WithPlugin(loadvariationriskbalancing.Name, loadvariationriskbalancing.New),
		app.WithPlugin(noderesources.AllocatableName, noderesources.NewAllocatable),
		app.WithPlugin(noderesourcetopology.Name, noderesourcetopology.New),
		app.WithPlugin(preemptiontoleration.Name, preemptiontoleration.New),
		app.WithPlugin(targetloadpacking.Name, targetloadpacking.New),
		// Sample plugins below.
		// app.WithPlugin(crossnodepreemption.Name, crossnodepreemption.New),
		app.WithPlugin(podstate.Name, podstate.New),
		app.WithPlugin(qos.Name, qos.New),
		app.WithPlugin(shmscoring.Name, shmscoring.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
