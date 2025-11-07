package main

import (
	"context"
	"flag"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/gzip"
	hertzPrometheus "github.com/hertz-contrib/monitor-prometheus"
	hertzlogrus "github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
	"github.com/hertz-contrib/obs-opentelemetry/provider"
	"github.com/hertz-contrib/obs-opentelemetry/tracing"
	"github.com/hertz-contrib/pprof"

	"github.com/jacob-Ho123/hertz-gateway-consul/common"
	"github.com/jacob-Ho123/hertz-gateway-consul/common/metric"
	"github.com/jacob-Ho123/hertz-gateway-consul/middleware"
)

func main() {
	addr := flag.String("addr", ":8080", "proxy server address")
	configFile := flag.String("config", "./config/config.yaml", "path to config file (JSON or YAML)")
	flag.Parse()
	hlog.SetLogger(hertzlogrus.NewLogger())
	hlog.SetLevel(hlog.LevelDebug)
	// load config
	config, err := common.LoadConfig(*configFile)
	if err != nil {
		hlog.Fatalf("Error loading  config: %v", err)
	}
	// init tracer
	if config.Common.OTLPEndpoint != "" {
		p := provider.NewOpenTelemetryProvider(
			provider.WithServiceName("hertz-gateway"),
			provider.WithExportEndpoint(config.Common.OTLPEndpoint),
			provider.WithEnableMetrics(false),
			provider.WithInsecure(),
		)
		defer p.Shutdown(context.Background())
	}
	tracer, cfg := tracing.NewServerTracer()
	// init metrics
	promHttpAddr := "0.0.0.0:9100"
	if config.Common.MetricsAddr != "" {
		promHttpAddr = config.Common.MetricsAddr
	}
	promMetricsPath := "/metrics"
	if config.Common.MetricsPath != "" {
		promMetricsPath = config.Common.MetricsPath
	}
	metric.InitPrometheus()
	h := server.Default(
		server.WithHostPorts(*addr),
		// prometheus
		server.WithTracer(hertzPrometheus.NewServerTracer(
			promHttpAddr,
			promMetricsPath,
			hertzPrometheus.WithRegistry(metric.GetRegistry()),
			hertzPrometheus.WithEnableGoCollector(true),
		)),
		tracer,
	)
	h.Use(gzip.Gzip(gzip.DefaultCompression))
	h.Use(tracing.ServerMiddleware(cfg))
	h.Use(recovery.Recovery(recovery.WithRecoveryHandler(common.RecoveryHandler)))
	// replace req or resp with config
	h.Use(middleware.MappingMiddleware(config))
	h.Use(middleware.ReverseProxyMiddleware(config))
	if os.Getenv("PROF") == "true" {
		pprof.Register(h)
	}
	h.Any("/*path", func(ctx context.Context, c *app.RequestContext) {
		// check if the path is valid
		if c.IsAborted() {
			return
		}
		hlog.CtxInfof(ctx, "invalid path: %s", c.Request.Path())
	})
	hlog.Infof("Starting proxy server on %s", *addr)
	h.Spin()
}
