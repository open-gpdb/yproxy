# yproxy

yproxy - a service for efficient transfer data external storages.


## debugging

1. set `debug_port` and `debug_minutes` in configuration file
2. send SIGHUP to `yproxy`
3. target golang tools to /debug/pprof to profile application

## metrics

yproxy gather and report various internal metrics in a prometheus format.
To get metrics query yproxy use http, default metrics port is 2112, `curl` example:

```
curl localhost:2112/metrics
```

The output consists of 2 categories of metrics:

1. Default go metrics:

- Goroutines and Threads: `go_goroutines` (current number of goroutines) and `go_threads` (number of OS threads).
- Garbage Collection: `go_gc_*` provides a summary/histogram of GC internals.
- Go Environment: `go_info` gives details about the Go version.
- Memory Statistics (Memstats): A comprehensive set of metrics tracking memory allocation and usage, including `go_memstats_alloc_bytes`

2. Custom yproxy metrics:

| Category  | Name | Type | Description |
|-----------|------|------|-------------|
| Connections | `client_connections` | gauge | The number of client (yezzey) connections to yproxy |
|             | `external_connections_backup` | gauge | The number of external yproxy connections to S3 backup storage |
|             | `external_connections_yezzey` | gauge | The number of external yproxy connections to S3 yezzey storage |
| Requests    | `read_req_processed_total` | counter | The total number of processed read (download) requests to S3 storage |
|             | `read_req_errors_total` | counter | The total number of errors occurred while processing read requests |
|             | `write_req_processed_total` | counter | The total number of processed write (upload) requests to S3 storage | 
|             | `write_req_errors_total` | counter | The total number of errors occurred while processing write requests |
| Internal    | `request_latency_bucket` | histogram | The number of requests and requests time for each source request |
|             | `request_size_bucket` | histogram | The number of requests and requests size for each source request |

Internal metrics show latency and size for each source request. Source request is the internal type of query from GP to yproxy. Yproxy performs mostly upload/download requests to S3. But yezzey perform requests multiple species. They are named as source request. We measure source requests:
```
READ
WRITE
LIMIT_READ
LIMIT_WRITE
S3_PUT
S3_GET
CAT
CATV2
PUT
PUTV2
PUTV3
DELETE
LIST
LISTV2
OBJECT META
COPY
COPYV2
GOOL
ERROR
UNTRASHIFY
COLLECT OBSOLETE
DELETE OBSOLETE
```

<details>
  <summary>Example of gathered metrics for a simple yezzey query:</summary>

```
query:
test=# select count(*) from public.ydt_100m;
   count
-----------
 100000000

yproxy metrics:
# HELP client_connections The number of client connections to yproxy
# TYPE client_connections gauge
client_connections 0
# HELP external_connections_backup The number of external connections to S3 storage
# TYPE external_connections_backup gauge
external_connections_backup 0
# HELP external_connections_yezzey The number of external connections to S3 storage
# TYPE external_connections_yezzey gauge
external_connections_yezzey 0
# HELP go_gc_duration_seconds A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 9.0478e-05
go_gc_duration_seconds{quantile="0.25"} 0.000105701
go_gc_duration_seconds{quantile="0.5"} 0.00013717
go_gc_duration_seconds{quantile="0.75"} 0.000170231
go_gc_duration_seconds{quantile="1"} 0.000334833
go_gc_duration_seconds_sum 0.00110412
go_gc_duration_seconds_count 7
# HELP go_gc_gogc_percent Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent.
# TYPE go_gc_gogc_percent gauge
go_gc_gogc_percent 100
# HELP go_gc_gomemlimit_bytes Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes.
# TYPE go_gc_gomemlimit_bytes gauge
go_gc_gomemlimit_bytes 2.5769803776e+10
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.24.10"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 4.430048e+06
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 1.4421808e+07
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.451385e+06
# HELP go_memstats_frees_total Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 75601
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 3.165672e+06
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 4.430048e+06
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 8.953856e+06
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 6.807552e+06
# HELP go_memstats_heap_objects Number of currently allocated objects. Equals to /gc/heap/objects:objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 22031
# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.
# TYPE go_memstats_heap_released_bytes gauge
go_memstats_heap_released_bytes 6.504448e+06
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 1.5761408e+07
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 1.7685566443842454e+09
# HELP go_memstats_mallocs_total Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 97632
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 38656
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 47112
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 225600
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 277440
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 7.28113e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 3.010015e+06
# HELP go_memstats_stack_inuse_bytes Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 1.015808e+06
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 1.015808e+06
# HELP go_memstats_sys_bytes Number of bytes obtained from system. Equals to /memory/classes/total:byte.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 2.472884e+07
# HELP go_sched_gomaxprocs_threads The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads.
# TYPE go_sched_gomaxprocs_threads gauge
go_sched_gomaxprocs_threads 32
# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads 19
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0.98
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 524288
# HELP process_network_receive_bytes_total Number of bytes received by the process over the network.
# TYPE process_network_receive_bytes_total counter
process_network_receive_bytes_total 1.553679757601e+12
# HELP process_network_transmit_bytes_total Number of bytes sent by the process over the network.
# TYPE process_network_transmit_bytes_total counter
process_network_transmit_bytes_total 1.520048543973e+12
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 13
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 2.9097984e+07
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.76855637857e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 2.56823296e+09
# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge
process_virtual_memory_max_bytes 1.8446744073709552e+19
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 2
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP read_req_errors_total The total number of errors during reads
# TYPE read_req_errors_total counter
read_req_errors_total 0
# HELP read_req_processed_total The total number of processed reads
# TYPE read_req_processed_total counter
read_req_processed_total 7183
# HELP request_latency Request latency in seconds
# TYPE request_latency histogram
request_latency_bucket{source="CATV2",le="0.001"} 0
request_latency_bucket{source="CATV2",le="0.01"} 0
request_latency_bucket{source="CATV2",le="0.05"} 0
request_latency_bucket{source="CATV2",le="0.1"} 0
request_latency_bucket{source="CATV2",le="0.5"} 0
request_latency_bucket{source="CATV2",le="1"} 0
request_latency_bucket{source="CATV2",le="5"} 0
request_latency_bucket{source="CATV2",le="10"} 0
request_latency_bucket{source="CATV2",le="50"} 4
request_latency_bucket{source="CATV2",le="100"} 4
request_latency_bucket{source="CATV2",le="500"} 4
request_latency_bucket{source="CATV2",le="1000"} 4
request_latency_bucket{source="CATV2",le="+Inf"} 4
request_latency_sum{source="CATV2"} 54.54297553986896
request_latency_count{source="CATV2"} 4
request_latency_bucket{source="READ",le="0.001"} 0
request_latency_bucket{source="READ",le="0.01"} 0
request_latency_bucket{source="READ",le="0.05"} 139
request_latency_bucket{source="READ",le="0.1"} 218
request_latency_bucket{source="READ",le="0.5"} 4537
request_latency_bucket{source="READ",le="1"} 6180
request_latency_bucket{source="READ",le="5"} 6651
request_latency_bucket{source="READ",le="10"} 6704
request_latency_bucket{source="READ",le="50"} 6821
request_latency_bucket{source="READ",le="100"} 6942
request_latency_bucket{source="READ",le="500"} 6978
request_latency_bucket{source="READ",le="1000"} 6984
request_latency_bucket{source="READ",le="+Inf"} 7183
request_latency_sum{source="READ"} 2.4379078546498725e+06
request_latency_count{source="READ"} 7183
request_latency_bucket{source="S3_GET",le="0.001"} 0
request_latency_bucket{source="S3_GET",le="0.01"} 0
request_latency_bucket{source="S3_GET",le="0.05"} 0
request_latency_bucket{source="S3_GET",le="0.1"} 0
request_latency_bucket{source="S3_GET",le="0.5"} 0
request_latency_bucket{source="S3_GET",le="1"} 0
request_latency_bucket{source="S3_GET",le="5"} 4
request_latency_bucket{source="S3_GET",le="10"} 4
request_latency_bucket{source="S3_GET",le="50"} 4
request_latency_bucket{source="S3_GET",le="100"} 4
request_latency_bucket{source="S3_GET",le="500"} 4
request_latency_bucket{source="S3_GET",le="1000"} 4
request_latency_bucket{source="S3_GET",le="+Inf"} 4
request_latency_sum{source="S3_GET"} 11.616734938160908
request_latency_count{source="S3_GET"} 4
# HELP request_size Request latency in seconds
# TYPE request_size histogram
request_size_bucket{source="CATV2",le="1"} 0
request_size_bucket{source="CATV2",le="128"} 0
request_size_bucket{source="CATV2",le="1024"} 0
request_size_bucket{source="CATV2",le="131072"} 0
request_size_bucket{source="CATV2",le="1.048576e+06"} 0
request_size_bucket{source="CATV2",le="2.097152e+06"} 0
request_size_bucket{source="CATV2",le="8.388608e+06"} 0
request_size_bucket{source="CATV2",le="1.6777216e+07"} 0
request_size_bucket{source="CATV2",le="1.34217728e+08"} 4
request_size_bucket{source="CATV2",le="1.073741824e+09"} 4
request_size_bucket{source="CATV2",le="+Inf"} 4
request_size_sum{source="CATV2"} 2.00261268e+08
request_size_count{source="CATV2"} 4
request_size_bucket{source="READ",le="1"} 412
request_size_bucket{source="READ",le="128"} 824
request_size_bucket{source="READ",le="1024"} 850
request_size_bucket{source="READ",le="131072"} 7183
request_size_bucket{source="READ",le="1.048576e+06"} 7183
request_size_bucket{source="READ",le="2.097152e+06"} 7183
request_size_bucket{source="READ",le="8.388608e+06"} 7183
request_size_bucket{source="READ",le="1.6777216e+07"} 7183
request_size_bucket{source="READ",le="1.34217728e+08"} 7183
request_size_bucket{source="READ",le="1.073741824e+09"} 7183
request_size_bucket{source="READ",le="+Inf"} 7183
request_size_sum{source="READ"} 2.00261356e+08
request_size_count{source="READ"} 7183
request_size_bucket{source="S3_GET",le="1"} 0
request_size_bucket{source="S3_GET",le="128"} 0
request_size_bucket{source="S3_GET",le="1024"} 0
request_size_bucket{source="S3_GET",le="131072"} 0
request_size_bucket{source="S3_GET",le="1.048576e+06"} 0
request_size_bucket{source="S3_GET",le="2.097152e+06"} 0
request_size_bucket{source="S3_GET",le="8.388608e+06"} 0
request_size_bucket{source="S3_GET",le="1.6777216e+07"} 0
request_size_bucket{source="S3_GET",le="1.34217728e+08"} 4
request_size_bucket{source="S3_GET",le="1.073741824e+09"} 4
request_size_bucket{source="S3_GET",le="+Inf"} 4
request_size_sum{source="S3_GET"} 2.00261356e+08
request_size_count{source="S3_GET"} 4
# HELP write_req_errors_total The total number of errors during reads
# TYPE write_req_errors_total counter
write_req_errors_total 0
# HELP write_req_processed_total The total number of processed writes
# TYPE write_req_processed_total counter
write_req_processed_total 0
[COMPUTE-PROD]root@rc1b-qfkbl5b84p1r8te6 ~ # curl localhost:2112/metrics
# HELP client_connections The number of client connections to yproxy
# TYPE client_connections gauge
client_connections 0
# HELP external_connections_backup The number of external connections to S3 storage
# TYPE external_connections_backup gauge
external_connections_backup 0
# HELP external_connections_yezzey The number of external connections to S3 storage
# TYPE external_connections_yezzey gauge
external_connections_yezzey 0
# HELP go_gc_duration_seconds A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 9.0478e-05
go_gc_duration_seconds{quantile="0.25"} 0.000105701
go_gc_duration_seconds{quantile="0.5"} 0.000130359
go_gc_duration_seconds{quantile="0.75"} 0.000147236
go_gc_duration_seconds{quantile="1"} 0.000334833
go_gc_duration_seconds_sum 0.001461997
go_gc_duration_seconds_count 10
# HELP go_gc_gogc_percent Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent.
# TYPE go_gc_gogc_percent gauge
go_gc_gogc_percent 100
# HELP go_gc_gomemlimit_bytes Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes.
# TYPE go_gc_gomemlimit_bytes gauge
go_gc_gomemlimit_bytes 2.5769803776e+10
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 14
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.24.10"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 3.966824e+06
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 2.0728512e+07
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.452297e+06
# HELP go_memstats_frees_total Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 116797
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 3.175912e+06
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 3.966824e+06
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 9.412608e+06
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 6.316032e+06
# HELP go_memstats_heap_objects Number of currently allocated objects. Equals to /gc/heap/objects:objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 19398
# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.
# TYPE go_memstats_heap_released_bytes gauge
go_memstats_heap_released_bytes 6.18496e+06
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 1.572864e+07
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 1.7685567256385603e+09
# HELP go_memstats_mallocs_total Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 136195
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 38656
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 47112
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 221120
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 277440
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 7.339234e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 2.998863e+06
# HELP go_memstats_stack_inuse_bytes Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 1.048576e+06
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 1.048576e+06
# HELP go_memstats_sys_bytes Number of bytes obtained from system. Equals to /memory/classes/total:byte.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 2.472884e+07
# HELP go_sched_gomaxprocs_threads The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads.
# TYPE go_sched_gomaxprocs_threads gauge
go_sched_gomaxprocs_threads 32
# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads 20
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 1.9
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 524288
# HELP process_network_receive_bytes_total Number of bytes received by the process over the network.
# TYPE process_network_receive_bytes_total counter
process_network_receive_bytes_total 1.553681055589e+12
# HELP process_network_transmit_bytes_total Number of bytes sent by the process over the network.
# TYPE process_network_transmit_bytes_total counter
process_network_transmit_bytes_total 1.520049112935e+12
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 13
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 2.9564928e+07
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.76855637857e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 2.643734528e+09
# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge
process_virtual_memory_max_bytes 1.8446744073709552e+19
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 3
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP read_req_errors_total The total number of errors during reads
# TYPE read_req_errors_total counter
read_req_errors_total 0
# HELP read_req_processed_total The total number of processed reads
# TYPE read_req_processed_total counter
read_req_processed_total 14362
# HELP request_latency Request latency in seconds
# TYPE request_latency histogram
request_latency_bucket{source="CATV2",le="0.001"} 0
request_latency_bucket{source="CATV2",le="0.01"} 0
request_latency_bucket{source="CATV2",le="0.05"} 0
request_latency_bucket{source="CATV2",le="0.1"} 0
request_latency_bucket{source="CATV2",le="0.5"} 0
request_latency_bucket{source="CATV2",le="1"} 0
request_latency_bucket{source="CATV2",le="5"} 0
request_latency_bucket{source="CATV2",le="10"} 0
request_latency_bucket{source="CATV2",le="50"} 8
request_latency_bucket{source="CATV2",le="100"} 8
request_latency_bucket{source="CATV2",le="500"} 8
request_latency_bucket{source="CATV2",le="1000"} 8
request_latency_bucket{source="CATV2",le="+Inf"} 8
request_latency_sum{source="CATV2"} 102.52228865142972
request_latency_count{source="CATV2"} 8
request_latency_bucket{source="READ",le="0.001"} 0
request_latency_bucket{source="READ",le="0.01"} 0
request_latency_bucket{source="READ",le="0.05"} 259
request_latency_bucket{source="READ",le="0.1"} 417
request_latency_bucket{source="READ",le="0.5"} 8791
request_latency_bucket{source="READ",le="1"} 12354
request_latency_bucket{source="READ",le="5"} 13293
request_latency_bucket{source="READ",le="10"} 13406
request_latency_bucket{source="READ",le="50"} 13623
request_latency_bucket{source="READ",le="100"} 13885
request_latency_bucket{source="READ",le="500"} 13953
request_latency_bucket{source="READ",le="1000"} 13967
request_latency_bucket{source="READ",le="+Inf"} 14362
request_latency_sum{source="READ"} 4.391004465408538e+06
request_latency_count{source="READ"} 14362
request_latency_bucket{source="S3_GET",le="0.001"} 0
request_latency_bucket{source="S3_GET",le="0.01"} 0
request_latency_bucket{source="S3_GET",le="0.05"} 0
request_latency_bucket{source="S3_GET",le="0.1"} 0
request_latency_bucket{source="S3_GET",le="0.5"} 0
request_latency_bucket{source="S3_GET",le="1"} 1
request_latency_bucket{source="S3_GET",le="5"} 8
request_latency_bucket{source="S3_GET",le="10"} 8
request_latency_bucket{source="S3_GET",le="50"} 8
request_latency_bucket{source="S3_GET",le="100"} 8
request_latency_bucket{source="S3_GET",le="500"} 8
request_latency_bucket{source="S3_GET",le="1000"} 8
request_latency_bucket{source="S3_GET",le="+Inf"} 8
request_latency_sum{source="S3_GET"} 17.627008362700515
request_latency_count{source="S3_GET"} 8
# HELP request_size Request latency in seconds
# TYPE request_size histogram
request_size_bucket{source="CATV2",le="1"} 0
request_size_bucket{source="CATV2",le="128"} 0
request_size_bucket{source="CATV2",le="1024"} 0
request_size_bucket{source="CATV2",le="131072"} 0
request_size_bucket{source="CATV2",le="1.048576e+06"} 0
request_size_bucket{source="CATV2",le="2.097152e+06"} 0
request_size_bucket{source="CATV2",le="8.388608e+06"} 0
request_size_bucket{source="CATV2",le="1.6777216e+07"} 0
request_size_bucket{source="CATV2",le="1.34217728e+08"} 8
request_size_bucket{source="CATV2",le="1.073741824e+09"} 8
request_size_bucket{source="CATV2",le="+Inf"} 8
request_size_sum{source="CATV2"} 4.00522536e+08
request_size_count{source="CATV2"} 8
request_size_bucket{source="READ",le="1"} 824
request_size_bucket{source="READ",le="128"} 1648
request_size_bucket{source="READ",le="1024"} 1691
request_size_bucket{source="READ",le="131072"} 14362
request_size_bucket{source="READ",le="1.048576e+06"} 14362
request_size_bucket{source="READ",le="2.097152e+06"} 14362
request_size_bucket{source="READ",le="8.388608e+06"} 14362
request_size_bucket{source="READ",le="1.6777216e+07"} 14362
request_size_bucket{source="READ",le="1.34217728e+08"} 14362
request_size_bucket{source="READ",le="1.073741824e+09"} 14362
request_size_bucket{source="READ",le="+Inf"} 14362
request_size_sum{source="READ"} 4.00522712e+08
request_size_count{source="READ"} 14362
request_size_bucket{source="S3_GET",le="1"} 0
request_size_bucket{source="S3_GET",le="128"} 0
request_size_bucket{source="S3_GET",le="1024"} 0
request_size_bucket{source="S3_GET",le="131072"} 0
request_size_bucket{source="S3_GET",le="1.048576e+06"} 0
request_size_bucket{source="S3_GET",le="2.097152e+06"} 0
request_size_bucket{source="S3_GET",le="8.388608e+06"} 0
request_size_bucket{source="S3_GET",le="1.6777216e+07"} 0
request_size_bucket{source="S3_GET",le="1.34217728e+08"} 8
request_size_bucket{source="S3_GET",le="1.073741824e+09"} 8
request_size_bucket{source="S3_GET",le="+Inf"} 8
request_size_sum{source="S3_GET"} 4.00522712e+08
request_size_count{source="S3_GET"} 8
# HELP write_req_errors_total The total number of errors during reads
# TYPE write_req_errors_total counter
write_req_errors_total 0
# HELP write_req_processed_total The total number of processed writes
# TYPE write_req_processed_total counter
write_req_processed_total 0
```
</details>
