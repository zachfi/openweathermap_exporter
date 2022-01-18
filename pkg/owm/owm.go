package owm

import (
	"net/http"

	"github.com/go-kit/log"
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

	return o, nil
}

func (o *OWM) Run() error {
	d := http.NewServeMux()
	d.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(o.cfg.ListenAddr, d)
}
