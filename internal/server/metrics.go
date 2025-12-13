package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// InitMetrics 初始化 OpenTelemetry Prometheus 指标导出器
// 采集的指标包括：
// - server_requests_total: 请求总数计数器
// - server_requests_duration_seconds: 请求延迟直方图
//
// 标签说明：
// - kind: 请求类型 (HTTP 或 gRPC)
// - operation: 操作名称 (如 /api.health.v1.Health/Ping)
// - path: HTTP 请求路径 (如 /ping)，gRPC 为空
// - method: HTTP 请求方法 (如 GET, POST)，gRPC 为空
// - code: 状态码 (0 为成功)
// - reason: 错误原因
func InitMetrics() error {
	exporter, err := otelprometheus.New(
		otelprometheus.WithRegisterer(prometheus.DefaultRegisterer),
	)
	if err != nil {
		return err
	}

	// 自定义直方图 bucket 配置
	histogramView := sdkmetric.NewView(
		sdkmetric.Instrument{Name: "server_requests_duration_seconds"},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
		},
	)

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(histogramView),
	)
	otel.SetMeterProvider(provider)

	return nil
}
