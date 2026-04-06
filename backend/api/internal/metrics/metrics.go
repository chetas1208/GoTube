package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gotube_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gotube_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	VideosProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gotube_videos_processed_total",
		Help: "Total number of videos processed",
	}, []string{"status"})

	VideoProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "gotube_video_processing_duration_seconds",
		Help:    "Video processing duration in seconds",
		Buckets: []float64{5, 10, 30, 60, 120, 300, 600},
	})

	QueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotube_queue_depth",
		Help: "Current number of messages in the job queue",
	})

	UploadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gotube_uploads_total",
		Help: "Total number of upload operations",
	}, []string{"status"})
)
