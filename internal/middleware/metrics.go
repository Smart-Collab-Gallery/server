package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	metricLabelKind      = "kind"
	metricLabelOperation = "operation"
	metricLabelPath      = "path"
	metricLabelMethod    = "method"
	metricLabelCode      = "code"
	metricLabelReason    = "reason"
)

// MetricsServer 自定义服务端指标中间件
// 相比 Kratos 默认的 metrics.Server()，增加了 path 和 method 标签
func MetricsServer() middleware.Middleware {
	meter := otel.Meter("server")

	requestsCounter, _ := meter.Int64Counter(
		"server_requests_total",
		metric.WithDescription("Total number of requests"),
		metric.WithUnit("{call}"),
	)

	secondsHistogram, _ := meter.Float64Histogram(
		"server_requests_duration_seconds",
		metric.WithDescription("Request duration in seconds"),
		metric.WithUnit("s"),
	)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				kind      string
				operation string
				path      string
				method    string
			)

			startTime := time.Now()

			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()

				// 如果是 HTTP 请求，获取 path 和 method
				if httpTr, ok := info.(*http.Transport); ok {
					if httpTr.Request() != nil {
						path = httpTr.Request().URL.Path
						method = httpTr.Request().Method
					}
				}
			}

			// 执行请求
			reply, err := handler(ctx, req)

			// 计算耗时
			duration := time.Since(startTime).Seconds()

			// 获取状态码和错误原因
			var code int
			var reason string
			if se := errors.FromError(err); se != nil {
				code = int(se.Code)
				reason = se.Reason
			}

			// 构建标签
			attrs := []attribute.KeyValue{
				attribute.String(metricLabelKind, kind),
				attribute.String(metricLabelOperation, operation),
				attribute.String(metricLabelPath, path),
				attribute.String(metricLabelMethod, method),
				attribute.Int(metricLabelCode, code),
				attribute.String(metricLabelReason, reason),
			}

			// 记录指标
			requestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
			secondsHistogram.Record(ctx, duration, metric.WithAttributes(attrs...))

			return reply, err
		}
	}
}
