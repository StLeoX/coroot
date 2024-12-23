package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/chpool"
	chproto "github.com/ClickHouse/ch-go/proto"
	"golang.org/x/exp/maps"
)

const (
	ttlDays = "7"
	// short time cache
	ttlHours = "2"
)

// 使用 Clickhouse 集群环境的条件：1.存在 system.zookeeper 配置。2.集群名称为 coroot 或者 default。
// 单机环境下，可能 SHOW CLUSTERS 返回 default，但是 EXISTS system.zookeeper 返回 0。
func getCluster(ctx context.Context, chPool *chpool.Pool) (string, error) {
	var exists chproto.ColUInt8
	q := ch.Query{Body: "EXISTS system.zookeeper", Result: chproto.Results{{Name: "result", Data: &exists}}}
	if err := chPool.Do(ctx, q); err != nil {
		return "", err
	}
	if exists.Row(0) != 1 {
		return "", nil
	}
	var clusterCol chproto.ColStr
	clusters := map[string]bool{}
	q = ch.Query{
		Body: "SHOW CLUSTERS",
		Result: chproto.Results{
			{Name: "cluster", Data: &clusterCol},
		},
		OnResult: func(ctx context.Context, block chproto.Block) error {
			return clusterCol.ForEach(func(i int, s string) error {
				clusters[s] = true
				return nil
			})
		},
	}
	if err := chPool.Do(ctx, q); err != nil {
		return "", err
	}
	switch {
	case len(clusters) == 0:
		return "", nil
	case len(clusters) == 1:
		return maps.Keys(clusters)[0], nil
	case clusters["coroot"]:
		return "coroot", nil
	case clusters["default"]:
		return "default", nil
	}
	return "", fmt.Errorf(`multiple ClickHouse clusters found, but neither "coroot" nor "default" cluster found`)
}

func (c *Collector) migrate(ctx context.Context, client *chClient) error {
	for _, t := range tables {
		t = strings.ReplaceAll(t, "@ttl_days", ttlDays)
		t = strings.ReplaceAll(t, "@ttl_hours", ttlHours)
		if client.cluster != "" {
			t = strings.ReplaceAll(t, "@on_cluster", "ON CLUSTER "+client.cluster)
			t = strings.ReplaceAll(t, "@merge_tree", "ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/{table}', '{replica}')")
			t = strings.ReplaceAll(t, "@replacing_merge_tree", "ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/{database}/{table}', '{replica}')")
		} else {
			t = strings.ReplaceAll(t, "@on_cluster", "")
			t = strings.ReplaceAll(t, "@merge_tree", "MergeTree()")
			t = strings.ReplaceAll(t, "@replacing_merge_tree", "ReplacingMergeTree()")
		}
		var result chproto.Results
		err := client.pool.Do(ctx, ch.Query{
			Body: t,
			OnResult: func(ctx context.Context, block chproto.Block) error {
				return nil
			},
			Result: result.Auto(),
		})
		if err != nil {
			return err
		}
	}
	if client.cluster != "" {
		for _, t := range distributedTables {
			t = strings.ReplaceAll(t, "@cluster", client.cluster)
			var result chproto.Results
			err := client.pool.Do(ctx, ch.Query{
				Body: t,
				OnResult: func(ctx context.Context, block chproto.Block) error {
					return nil
				},
				Result: result.Auto(),
			})
			if err != nil {
				return err
			}
		}

	}
	return nil
}

var (
	tables = []string{
		// 新建表 otel_logs。
		`
CREATE TABLE IF NOT EXISTS otel_logs @on_cluster (
     Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
     TraceId String CODEC(ZSTD(1)),
     SpanId String CODEC(ZSTD(1)),
     TraceFlags UInt32 CODEC(ZSTD(1)),
     SeverityText LowCardinality(String) CODEC(ZSTD(1)),
     SeverityNumber Int32 CODEC(ZSTD(1)),
     ServiceName LowCardinality(String) CODEC(ZSTD(1)),
     Body String CODEC(ZSTD(1)),
     ResourceAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     LogAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     INDEX idx_trace_id TraceId TYPE bloom_filter(0.001) GRANULARITY 1,
     INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_res_attr_value mapValues(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_log_attr_key mapKeys(LogAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_log_attr_value mapValues(LogAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_body Body TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1
) ENGINE @merge_tree
TTL toDateTime(Timestamp) + toIntervalDay(@ttl_days)
PARTITION BY toDate(Timestamp)
ORDER BY (ServiceName, SeverityText, toUnixTimestamp(Timestamp))
SETTINGS index_granularity=8192, ttl_only_drop_parts = 1
`,

		// 新建表 otel_traces。
		`
CREATE TABLE IF NOT EXISTS otel_traces @on_cluster (
     Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
     TraceId String CODEC(ZSTD(1)),
     SpanId String CODEC(ZSTD(1)),
     ParentSpanId String CODEC(ZSTD(1)),
     TraceState String CODEC(ZSTD(1)),
     SpanName LowCardinality(String) CODEC(ZSTD(1)),
     SpanKind LowCardinality(String) CODEC(ZSTD(1)),
     ServiceName LowCardinality(String) CODEC(ZSTD(1)),
     ResourceAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     SpanAttributes Map(LowCardinality(String), String) CODEC(ZSTD(1)),
     Duration Int64 CODEC(ZSTD(1)),
     StatusCode LowCardinality(String) CODEC(ZSTD(1)),
     StatusMessage String CODEC(ZSTD(1)),
     Events Nested (
         Timestamp DateTime64(9),
         Name LowCardinality(String),
         Attributes Map(LowCardinality(String), String)
     ) CODEC(ZSTD(1)),
     Links Nested (
         TraceId String,
         SpanId String,
         TraceState String,
         Attributes Map(LowCardinality(String), String)
     ) CODEC(ZSTD(1)),
     INDEX idx_trace_id TraceId TYPE bloom_filter(0.001) GRANULARITY 1,
     INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_res_attr_value mapValues(ResourceAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_span_attr_key mapKeys(SpanAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_span_attr_value mapValues(SpanAttributes) TYPE bloom_filter(0.01) GRANULARITY 1,
     INDEX idx_duration Duration TYPE minmax GRANULARITY 1
) ENGINE @merge_tree
TTL toDateTime(Timestamp) + toIntervalDay(@ttl_days)
PARTITION BY toDate(Timestamp)
ORDER BY (ServiceName, SpanName, toUnixTimestamp(Timestamp))
SETTINGS index_granularity=8192, ttl_only_drop_parts = 1`,

		// 新建物化列 NetSockPeerAddr。
		`
ALTER TABLE otel_traces ADD COLUMN IF NOT EXISTS NetSockPeerAddr LowCardinality(String) 
MATERIALIZED concat(SpanAttributes['net.peer.name'], ':', SpanAttributes['net.peer.port']) CODEC(ZSTD(1))`,

		// 新建 ebpf_server_spans 表
		`
CREATE TABLE IF NOT EXISTS ebpf_server_spans @on_cluster (
     Timestamp DateTime64(9) CODEC(Delta, ZSTD(1)),
     Duration Int64 CODEC(ZSTD(1)),
     ContainerId LowCardinality(String) CODEC(ZSTD(1)),
     TgidRead LowCardinality(String) CODEC(ZSTD(1)),
     TgidWrite LowCardinality(String) CODEC(ZSTD(1)),
     StatementId UInt32 CODEC(ZSTD(1))
) ENGINE @merge_tree
TTL toDateTime(Timestamp) + toIntervalHour(@ttl_hours)
PARTITION BY toDate(Timestamp)
ORDER BY (toUnixTimestamp(Timestamp))`,

		// 新建表 profiling_stacks。
		`
CREATE TABLE IF NOT EXISTS profiling_stacks @on_cluster (
	ServiceName LowCardinality(String) CODEC(ZSTD(1)),
	Hash UInt64 CODEC(ZSTD(1)),
	LastSeen DateTime64(9) CODEC(Delta, ZSTD(1)),
	Stack Array(String) CODEC(ZSTD(1))
) 
ENGINE @replacing_merge_tree
PRIMARY KEY (ServiceName, Hash)
TTL toDateTime(LastSeen) + toIntervalDay(@ttl_days)
PARTITION BY toDate(LastSeen)
ORDER BY (ServiceName, Hash)`,

		// 新建表 profiling_samples。
		`
CREATE TABLE IF NOT EXISTS profiling_samples @on_cluster (
	ServiceName LowCardinality(String) CODEC(ZSTD(1)),
    Type LowCardinality(String) CODEC(ZSTD(1)),
	Start DateTime64(9) CODEC(Delta, ZSTD(1)),
	End DateTime64(9) CODEC(Delta, ZSTD(1)),
	Labels Map(LowCardinality(String), String) CODEC(ZSTD(1)),
	StackHash UInt64 CODEC(ZSTD(1)),
	Value Int64 CODEC(ZSTD(1))
) ENGINE @merge_tree
TTL toDateTime(Start) + toIntervalDay(@ttl_days)
PARTITION BY toDate(Start)
ORDER BY (ServiceName, Type, toUnixTimestamp(Start), toUnixTimestamp(End))`,

		// 新建表 profiling_profiles。
		`
CREATE TABLE IF NOT EXISTS profiling_profiles @on_cluster (
    ServiceName LowCardinality(String) CODEC(ZSTD(1)),
    Type LowCardinality(String) CODEC(ZSTD(1)),
    LastSeen DateTime64(9) CODEC(Delta, ZSTD(1))
)
ENGINE @replacing_merge_tree
PRIMARY KEY (ServiceName, Type)
TTL toDateTime(LastSeen) + toIntervalDay(@ttl_days)
PARTITION BY toDate(LastSeen)`,

		// 新建物化视图 profiling_profiles_mv，为 profiling_profiles 维护 max range。
		`
CREATE MATERIALIZED VIEW IF NOT EXISTS profiling_profiles_mv @on_cluster TO profiling_profiles AS
SELECT ServiceName, Type, max(End) AS LastSeen FROM profiling_samples group by ServiceName, Type`,
	}

	distributedTables = []string{
		`CREATE TABLE IF NOT EXISTS otel_logs_distributed ON CLUSTER @cluster AS otel_logs
			ENGINE = Distributed(@cluster, currentDatabase(), otel_logs, rand())`,

		`CREATE TABLE IF NOT EXISTS otel_traces_distributed ON CLUSTER @cluster AS otel_traces
			ENGINE = Distributed(@cluster, currentDatabase(), otel_traces, cityHash64(TraceId))`,

		`CREATE TABLE IF NOT EXISTS ebpf_server_spans_distributed ON CLUSTER @cluster AS ebpf_server_spans
			ENGINE = Distributed(@cluster, currentDatabase(), ebpf_server_spans)`,

		`CREATE TABLE IF NOT EXISTS profiling_stacks_distributed ON CLUSTER @cluster AS profiling_stacks
		ENGINE = Distributed(@cluster, currentDatabase(), profiling_stacks, Hash)`,

		`CREATE TABLE IF NOT EXISTS profiling_samples_distributed ON CLUSTER @cluster AS profiling_samples
		ENGINE = Distributed(@cluster, currentDatabase(), profiling_samples, StackHash)`,

		`CREATE TABLE IF NOT EXISTS profiling_profiles_distributed ON CLUSTER @cluster AS profiling_profiles
		ENGINE = Distributed(@cluster, currentDatabase(), profiling_profiles)`,
	}
)

func ReplaceTables(query string, distributed bool) string {
	tbls := []string{"otel_logs", "otel_traces", "ebpf_server_spans", "profiling_stacks", "profiling_samples", "profiling_profiles"}
	for _, t := range tbls {
		placeholder := "@@table_" + t + "@@"
		if distributed {
			t += "_distributed"
		}
		query = strings.ReplaceAll(query, placeholder, t)
	}
	return query
}
