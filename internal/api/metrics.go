package api

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/metrics"
	"strings"
	"time"
)

// MetricsHandler returns an HTTP handler that exposes Go runtime telemetry
// and optional pipeline metrics in Prometheus text format.
func MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		var sb strings.Builder

		// --- Go runtime metrics using runtime/metrics (Go 1.16+) ---
		sb.WriteString("# HELP go_goroutines Number of goroutines that currently exist.\n")
		sb.WriteString("# TYPE go_goroutines gauge\n")
		sb.WriteString(fmt.Sprintf("go_goroutines %d\n", runtime.NumGoroutine()))

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		sb.WriteString("# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.\n")
		sb.WriteString("# TYPE go_memstats_alloc_bytes gauge\n")
		sb.WriteString(fmt.Sprintf("go_memstats_alloc_bytes %d\n", m.Alloc))

		sb.WriteString("# HELP go_memstats_heap_objects Number of allocated objects.\n")
		sb.WriteString("# TYPE go_memstats_heap_objects gauge\n")
		sb.WriteString(fmt.Sprintf("go_memstats_heap_objects %d\n", m.HeapObjects))

		sb.WriteString("# HELP go_memstats_total_alloc_bytes Total number of bytes allocated.\n")
		sb.WriteString("# TYPE go_memstats_total_alloc_bytes counter\n")
		sb.WriteString(fmt.Sprintf("go_memstats_total_alloc_bytes %d\n", m.TotalAlloc))

		sb.WriteString("# HELP go_memstats_sys_bytes Total bytes of memory obtained from the OS.\n")
		sb.WriteString("# TYPE go_memstats_sys_bytes gauge\n")
		sb.WriteString(fmt.Sprintf("go_memstats_sys_bytes %d\n", m.Sys))

		sb.WriteString("# HELP go_memstats_gc_cycles_total Number of completed GC cycles.\n")
		sb.WriteString("# TYPE go_memstats_gc_cycles_total counter\n")
		sb.WriteString(fmt.Sprintf("go_memstats_gc_cycles_total %d\n", m.NumGC))

		sb.WriteString("# HELP go_memstats_gc_pause_ns_total Total GC pause duration in nanoseconds.\n")
		sb.WriteString("# TYPE go_memstats_gc_pause_ns_total counter\n")
		sb.WriteString(fmt.Sprintf("go_memstats_gc_pause_ns_total %d\n", totalGCPause(&m)))

		// --- CPU count ---
		sb.WriteString("# HELP go_cpus The number of logical CPUs usable by the current process.\n")
		sb.WriteString("# TYPE go_cpus gauge\n")
		sb.WriteString(fmt.Sprintf("go_cpus %d\n", runtime.NumCPU()))

		// --- Read /gc/gomemlimit from runtime/metrics (Go 1.19+) ---
		sb.WriteString("# HELP go_mem_limit_max_bytes The maximum amount of memory the Go runtime can use.\n")
		sb.WriteString("# TYPE go_mem_limit_max_bytes gauge\n")
		readSample := []metrics.Sample{{Name: "/gc/gomemlimit:bytes"}}
		metrics.Read(readSample)
		if readSample[0].Value.Kind() != metrics.KindBad {
			sb.WriteString(fmt.Sprintf("go_mem_limit_max_bytes %d\n", readSample[0].Value.Uint64()))
		}

		w.Write([]byte(sb.String()))
	})
}

func totalGCPause(m *runtime.MemStats) uint64 {
	// PauseTotalNs is the cumulative nanoseconds in GC stop-the-world pauses
	return m.PauseTotalNs
}

var metricsStartTime = time.Now()

func init() {
	metricsStartTime = time.Now()
}
