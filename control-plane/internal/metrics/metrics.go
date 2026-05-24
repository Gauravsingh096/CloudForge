package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DeploysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cloudforge_deploys_total",
		Help: "Total deploys created, labelled by final status.",
	}, []string{"status"})

	DeployApplyDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cloudforge_deploy_apply_duration_seconds",
		Help:    "Seconds taken to kubectl-apply a deploy.",
		Buckets: prometheus.DefBuckets,
	}, []string{"app", "status"})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cloudforge_http_requests_total",
		Help: "Total HTTP requests handled by the control plane.",
	}, []string{"method", "path"})
)
