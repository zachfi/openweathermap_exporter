package owm

import (
	"context"
	"log"
	"strconv"

	owm "github.com/briandowns/openweathermap"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/otel/trace"
)

var (
	metricWeatherForecastDesc = prometheus.NewDesc(
		"weather_forecast",
		"Weather condition forecast",
		[]string{"location", "condition", "inhours"},
		nil,
	)

	metricWeatherCurrentDesc = prometheus.NewDesc(
		"weather_current",
		"Weather condition current",
		[]string{"location", "condition"},
		nil,
	)

	metricUVIndexCurrentDesc = prometheus.NewDesc(
		"uv_index_current",
		"Current UV Index",
		[]string{"location"},
		nil,
	)

	metricPollutionCurrentDesc = prometheus.NewDesc(
		"pollution_current",
		"Current Air Pollution",
		[]string{"location"},
		nil,
	)

	metricWeatherEpochDesc = prometheus.NewDesc(
		"weather_epoch",
		"Weather event, (sunrise|sunset)",
		[]string{"location", "event"},
		nil,
	)
)

func (o *OWM) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricWeatherForecastDesc
	ch <- metricWeatherCurrentDesc
	ch <- metricWeatherEpochDesc
	ch <- metricUVIndexCurrentDesc
	ch <- metricPollutionCurrentDesc
}

func (o *OWM) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	ctx, span := o.tracer.Start(ctx, "owmCollect")
	defer span.End()

	_ = level.Debug(o.logger).Log("msg", "collecting openweathermap data",
		"traceID", trace.SpanContextFromContext(ctx).TraceID().String(),
	)

	for _, location := range o.cfg.Locations {
		//
		// handle wind gust - Probably needs a PR upstream
		// handle UTC DT
		// handle feels_like - needs upstream PR
		o.collectWeatherForecast(ctx, ch, location)
		o.collectCurrentWeather(ctx, ch, location)
		o.collectPollution(ctx, ch, location)
	}

}

func (o *OWM) collectWeatherForecast(ctx context.Context, ch chan<- prometheus.Metric, location Location) {
	ctx, span := o.tracer.Start(ctx, "collectWeatherForecast")
	defer span.End()

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	w, err := owm.NewForecast("5", "C", "EN", o.cfg.APIKey, owm.WithHttpClient(o.client)) // valid options for first parameter are "5" and "16"
	if err != nil {
		_ = level.Error(o.logger).Log("msg", "failed to refresh forecast", "err", err)
		return
	}

	err = w.DailyByCoordinates(coord, 50)
	if err != nil {
		_ = level.Error(o.logger).Log("msg", "failed to fetch daily", "err", err)
		return
	}

	fore := w.ForecastWeatherJson.(*owm.Forecast5WeatherData)

	for i, p := range fore.List {
		inHours := strconv.Itoa(i * 3)

		conditions := map[string]float64{
			"clouds":      float64(p.Clouds.All),
			"humidity":    float64(p.Main.Humidity),
			"pressure":    p.Main.GrndLevel,
			"rain":        p.Rain.ThreeH,
			"snow":        p.Snow.ThreeH,
			"temp_max":    p.Main.TempMax,
			"temp_min":    p.Main.TempMin,
			"temp":        p.Main.Temp,
			"wind_degree": p.Wind.Deg,
			"wind_speed":  p.Wind.Speed,
		}

		for condition, value := range conditions {
			ch <- prometheus.MustNewConstMetric(
				metricWeatherForecastDesc,
				prometheus.GaugeValue,
				value,
				location.Name,
				condition,
				inHours,
			)
		}
	}
}

func (o *OWM) collectCurrentWeather(ctx context.Context, ch chan<- prometheus.Metric, location Location) {
	ctx, span := o.tracer.Start(ctx, "collectCurrentWeather")
	defer span.End()

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	c, err := owm.NewCurrent("C", "EN", o.cfg.APIKey, owm.WithHttpClient(o.client))
	if err != nil {
		_ = level.Error(o.logger).Log("msg", "failed to get new current client", "err", err)
	}

	err = c.CurrentByCoordinates(coord)
	if err != nil {
		_ = level.Error(o.logger).Log("msg", "failed to refresh current", "err", err)
		return
	}

	// sunrise := time.Unix(int64(c.Sys.Sunrise), 0)
	// _ = level.Info(o.logger).Log("msg", "current", "sunrise", fmt.Sprintf("%+v", sunrise.String()))

	ch <- prometheus.MustNewConstMetric(
		metricWeatherEpochDesc,
		prometheus.CounterValue,
		float64(c.Sys.Sunrise),
		location.Name,
		"sunrise",
	)

	// sunset := time.Unix(int64(c.Sys.Sunset), 0)
	// _ = level.Info(o.logger).Log("msg", "current", "sunset", fmt.Sprintf("%+v", sunset.String()))

	ch <- prometheus.MustNewConstMetric(
		metricWeatherEpochDesc,
		prometheus.CounterValue,
		float64(c.Sys.Sunset),
		location.Name,
		"sunset",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.Temp,
		location.Name,
		"temp",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.TempMax,
		location.Name,
		"temp_max",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.TempMin,
		location.Name,
		"temp_min",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.GrndLevel,
		location.Name,
		"pressure",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		float64(c.Main.Humidity),
		location.Name,
		"humidity",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Wind.Speed,
		location.Name,
		"wind_speed",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Wind.Deg,
		location.Name,
		"wind_degree",
	)

	// Send PR upstream for this
	// ch <- prometheus.MustNewConstMetric(
	// 	metricWeatherCurrentDesc,
	// 	prometheus.CounterValue,
	// 	float64(c.Wind.Gust),
	// 	location.Name,
	// 	"current_wind_gust",modules/inventory/network.go:
	// )

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Snow.ThreeH,
		location.Name,
		"snow_threeh",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Rain.ThreeH,
		location.Name,
		"rain_threeh",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		float64(c.Clouds.All),
		location.Name,
		"cloud_cover",
	)
}

func (o *OWM) collectPollution(ctx context.Context, ch chan<- prometheus.Metric, location Location) {
	ctx, span := o.tracer.Start(ctx, "collectPollution")
	defer span.End()

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	pollution, err := owm.NewPollution(o.cfg.APIKey, owm.WithHttpClient(o.client))
	if err != nil {
		_ = level.Error(o.logger).Log("msg", "failed to get new pollution data", "err", err)
	}

	params := &owm.PollutionParameters{
		Location: *coord,
		Datetime: "current",
	}

	if err := pollution.PollutionByParams(params); err != nil {
		log.Fatalln(err)
		_ = level.Error(o.logger).Log("msg", "failed to update pollution data", "err", err)
	}

	for _, p := range pollution.Data {
		ch <- prometheus.MustNewConstMetric(
			metricPollutionCurrentDesc,
			prometheus.GaugeValue,
			p.Value,
			location.Name,
			"pollution",
		)
	}

}
