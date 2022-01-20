package owm

import (
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xaque208/znet/pkg/util"
)

type OWM struct {
	cfg    Config
	logger log.Logger
}

func New(cfg Config) (*OWM, error) {
	o := &OWM{
		cfg: cfg,
	}

	o.logger = util.NewLogger()

	prometheus.MustRegister(o)

	return o, nil
}

func (o *OWM) Run() error {
	d := http.NewServeMux()
	d.Handle("/metrics", promhttp.Handler())

	_ = level.Info(o.logger).Log("msg", "openweathermap_exporter started")

	defer func() { _ = level.Info(o.logger).Log("msg", "openweathermap_exporter stopped") }()

	return http.ListenAndServe(o.cfg.ListenAddr, d)
}
