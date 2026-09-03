package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ProgressLogInterval is how often (in cumulative processed items) callers
// should emit an Info-level progress log line, to keep log volume bounded
// during long-running delete/untrashify operations.
const ProgressLogInterval = 50000

var (
	itemCountBuckets = []float64{1, 10, 50, 100, 500, 1000, 5000, 10000, 50000}

	DeleteProcessTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "delete_process_items",
		Help: "Size of the current batch of items found to process by delete/untrashify handlers",
	}, []string{"bucket", "operation"})

	DeleteProcessBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "delete_process_bytes",
		Help: "Total size in bytes of the current batch of items found to process by delete handlers",
	}, []string{"bucket", "operation"})

	DeleteProcessRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "delete_process_remaining",
		Help: "Items left to process by delete/untrashify handlers",
	}, []string{"bucket", "operation"})

	DeleteProcessRemainingBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "delete_process_remaining_bytes",
		Help: "Total size in bytes of items left to process by delete handlers",
	}, []string{"bucket", "operation"})

	DeleteProcessProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_processed_total",
		Help: "Cumulative number of items delete/untrashify handlers attempted to delete or move",
	}, []string{"bucket", "operation"})

	DeleteProcessMoved = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_moved_total",
		Help: "Cumulative number of garbage items successfully moved to trash",
	}, []string{"bucket", "operation"})

	DeleteProcessDeleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_deleted_total",
		Help: "Cumulative number of items successfully deleted from storage",
	}, []string{"bucket", "operation"})

	DeleteProcessKept = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_kept_total",
		Help: "Cumulative number of items intentionally left in place (skipped or failed after retries)",
	}, []string{"bucket", "operation"})

	DeleteProcessKeptBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_kept_bytes_total",
		Help: "Cumulative size in bytes of garbage objects skipped or left after failed retries",
	}, []string{"bucket", "operation"})

	DeleteProcessMovedBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_moved_bytes_total",
		Help: "Cumulative size in bytes of garbage objects successfully moved to trash",
	}, []string{"bucket", "operation"})

	DeleteProcessDeletedBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_process_deleted_bytes_total",
		Help: "Cumulative size in bytes of garbage objects successfully deleted from storage",
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
	processed atomic.Int64
}

func NewDeleteOpTracker(bucket, operation string) *DeleteOpTracker {
	return &DeleteOpTracker{bucket: bucket, operation: operation}
}

func (t *DeleteOpTracker) labels() prometheus.Labels {
	return prometheus.Labels{"bucket": t.bucket, "operation": t.operation}
}

func (t *DeleteOpTracker) Operation() string {
	return t.operation
}

// BeginScan starts reporting a new batch and clears the gauges left by the
// previous scan for the same bucket and operation. Call Discover once for each
// object classified as garbage after BeginScan.
func (t *DeleteOpTracker) BeginScan() {
	DeleteProcessTotal.With(t.labels()).Set(0)
	DeleteProcessBytes.With(t.labels()).Set(0)
	DeleteProcessRemaining.With(t.labels()).Set(0)
	DeleteProcessRemainingBytes.With(t.labels()).Set(0)
}

// Discover records one object classified as garbage. The object contributes to
// both the complete batch and its not-yet-processed part.
//
// Every Discover call may be followed by at most one of CompleteDeleted,
// CompleteMoved, or CompleteSkipped. Calling more than one completion method
// for the same object decrements the remaining gauges more than once.
func (t *DeleteOpTracker) Discover(size int64) {
	DeleteProcessTotal.With(t.labels()).Inc()
	DeleteProcessBytes.With(t.labels()).Add(float64(size))
	DeleteProcessRemaining.With(t.labels()).Inc()
	DeleteProcessRemainingBytes.With(t.labels()).Add(float64(size))
}

func (t *DeleteOpTracker) ObserveList(d time.Duration, itemCount int) {
	DeleteRequestLatency.With(
		prometheus.Labels{
			"bucket":    t.bucket,
			"operation": t.operation,
			"stage":     "list",
		},
	).Observe(d.Seconds())
	DeleteRequestSize.With(t.labels()).Observe(float64(itemCount))
}

func (t *DeleteOpTracker) ObserveDelete(d time.Duration) {
	DeleteRequestLatency.With(
		prometheus.Labels{
			"bucket":    t.bucket,
			"operation": t.operation,
			"stage":     "delete",
		},
	).Observe(d.Seconds())
}

// AddProcessed records n more processed items and returns the cumulative
// processed count for this tracker, for use with ProgressLogInterval.
func (t *DeleteOpTracker) AddProcessed(n int) int64 {
	DeleteProcessProcessed.With(t.labels()).Add(float64(n))
	return t.processed.Add(int64(n))
}

// CompleteDeleted records successful physical deletion of a discovered object
// and removes it from the remaining part of the batch.
func (t *DeleteOpTracker) CompleteDeleted(size int64) {
	DeleteProcessDeleted.With(t.labels()).Inc()
	DeleteProcessDeletedBytes.With(t.labels()).Add(float64(size))
	t.removePending(size)
}

// CompleteMoved records successful movement of a discovered object to trash
// and removes it from the remaining part of the batch.
func (t *DeleteOpTracker) CompleteMoved(size int64) {
	DeleteProcessMoved.With(t.labels()).Inc()
	DeleteProcessMovedBytes.With(t.labels()).Add(float64(size))
	t.removePending(size)
}

// RecordFailed records an object that remains after all processing retries.
// It intentionally does not remove the object from the remaining gauges.
func (t *DeleteOpTracker) RecordFailed(size int64) {
	DeleteProcessKept.With(t.labels()).Inc()
	DeleteProcessKeptBytes.With(t.labels()).Add(float64(size))
}

// CompleteSkipped records a discovered object intentionally excluded from
// processing and removes it from the remaining part of the batch.
func (t *DeleteOpTracker) CompleteSkipped(size int64) {
	t.RecordFailed(size)
	t.removePending(size)
}

func (t *DeleteOpTracker) removePending(size int64) {
	DeleteProcessRemaining.With(t.labels()).Dec()
	DeleteProcessRemainingBytes.With(t.labels()).Sub(float64(size))
}
