package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	itemCountBuckets = []float64{1, 10, 50, 100, 500, 1000, 5000, 10000, 50000}

	DeleteProcessTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "delete_process_total",
		Help: "Size of the current batch of items found to process by delete/untrashify handlers",
	}, []string{"bucket", "operation"})

	DeleteProcessRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "delete_process_remaining",
		Help: "Items left to process by delete/untrashify handlers",
	}, []string{"bucket", "operation"})

	DeleteProcessProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_processed_total",
		Help: "Cumulative number of items delete/untrashify handlers attempted to delete or move",
	}, []string{"bucket", "operation"})

	DeleteProcessDeleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_deleted_total",
		Help: "Cumulative number of items successfully deleted or moved",
	}, []string{"bucket", "operation"})

	DeleteProcessKept = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_kept_total",
		Help: "Cumulative number of items intentionally left in place (skipped or failed after retries)",
	}, []string{"bucket", "operation"})

	DeleteRequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "delete_request_latency_seconds",
		Help:    "Latency of delete-handler list/delete operations",
		Buckets: latencyBuckets,
	}, []string{"bucket", "operation", "stage"})

	DeleteRequestSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "delete_request_size",
		Help:    "Number of items returned by a delete-handler list operation",
		Buckets: itemCountBuckets,
	}, []string{"bucket", "operation"})
)

// DeleteOpTracker reports per-bucket, per-operation progress metrics for the
// long-running delete/untrashify handlers in pkg/proc/delete_handler.go.
type DeleteOpTracker struct {
	bucket    string
	operation string
}

func NewDeleteOpTracker(bucket, operation string) *DeleteOpTracker {
	return &DeleteOpTracker{bucket: bucket, operation: operation}
}

func (t *DeleteOpTracker) labels() prometheus.Labels {
	return prometheus.Labels{"bucket": t.bucket, "operation": t.operation}
}

func (t *DeleteOpTracker) SetTotal(n int) {
	DeleteProcessTotal.With(t.labels()).Set(float64(n))
}

func (t *DeleteOpTracker) SetRemaining(n int) {
	DeleteProcessRemaining.With(t.labels()).Set(float64(n))
}

func (t *DeleteOpTracker) ObserveList(d time.Duration, itemCount int) {
	DeleteRequestLatency.With(prometheus.Labels{"bucket": t.bucket, "operation": t.operation, "stage": "list"}).Observe(d.Seconds())
	DeleteRequestSize.With(t.labels()).Observe(float64(itemCount))
}

func (t *DeleteOpTracker) ObserveDelete(d time.Duration) {
	DeleteRequestLatency.With(prometheus.Labels{"bucket": t.bucket, "operation": t.operation, "stage": "delete"}).Observe(d.Seconds())
}

func (t *DeleteOpTracker) AddProcessed(n int) {
	DeleteProcessProcessed.With(t.labels()).Add(float64(n))
}

func (t *DeleteOpTracker) AddDeleted(n int) {
	DeleteProcessDeleted.With(t.labels()).Add(float64(n))
}

func (t *DeleteOpTracker) AddKept(n int) {
	DeleteProcessKept.With(t.labels()).Add(float64(n))
}
