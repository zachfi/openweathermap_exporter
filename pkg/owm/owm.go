package owm

import (
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xaque208/znet/pkg/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OWM struct {
	cfg    Config
	logger log.Logger
	tracer trace.Tracer
	client *http.Client
}

func New(cfg Config) (*OWM, error) {
	o := &OWM{
		cfg:    cfg,
		logger: util.NewLogger(),
		tracer: otel.Tracer("openWeatherMap"),
		client: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}

	prometheus.MustRegister(o)

	return o, nil
}

func (o *OWM) Run() error {
	d := http.NewServeMux()
	d.Handle("/metrics", promhttp.Handler())

	_ = level.Info(o.logger).Log("msg", fmt.Sprintf("openweathermap_exporter started on %s", o.cfg.ListenAddr))

	defer func() { _ = level.Info(o.logger).Log("msg", "openweathermap_exporter stopped") }()

	return http.ListenAndServe(o.cfg.ListenAddr, d)
}
