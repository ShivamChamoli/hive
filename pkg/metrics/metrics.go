/*
Copyright 2018 The Kubernetes Authors.

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

package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	hivev1 "github.com/openshift/hive/pkg/apis/hive/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	metricClusterDeploymentsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hive_cluster_deployments_total",
		Help: "Total number of cluster deployments that exist in Hive.",
	})
	metricClusterDeploymentsInstalledTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hive_cluster_deployments_installed_total",
		Help: "Total number of cluster deployments that are successfully installed.",
	})
)

// Calculator runs in a goroutine and periodically calculates and publishes
// Prometheus metrics which will be exposed at our /metrics endpoint.
// This should be used for metrics which do not fit well into controller reconcile loops,
// things that are calculated globally rather than metrics related to specific reconciliations.
type Calculator struct {
	Client client.Client

	// Interval is the length of time we sleep between metrics calculations.
	Interval time.Duration
}

// Run begins the metrics calculation loop.
func (mc *Calculator) Run() {

	// Run forever, sleep at the end:
	for {
		mcLog := log.WithField("component", "metrics")
		// Load all ClusterDeployments so we can accumulate facts about them.
		clusterDeployments := &hivev1.ClusterDeploymentList{}
		err := mc.Client.List(context.Background(), &client.ListOptions{}, clusterDeployments)
		if err != nil {
			log.WithError(err).Error("error listing cluster deployments")
		}
		mcLog.WithField("totalClusterDeployments", len(clusterDeployments.Items)).Debug("loaded cluster deployments")
		total := 0
		installedTotal := 0
		for _, cd := range clusterDeployments.Items {
			total = total + 1
			if cd.Status.Installed {
				installedTotal = installedTotal + 1
			}
		}
		metricClusterDeploymentsTotal.Set(float64(total))
		metricClusterDeploymentsInstalledTotal.Set(float64(installedTotal))

		time.Sleep(mc.Interval)
	}
}
