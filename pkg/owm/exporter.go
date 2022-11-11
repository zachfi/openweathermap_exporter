package owm

import (
	"context"
	"fmt"
	"strconv"
	"time"

	owm "github.com/briandowns/openweathermap"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	metricWeatherForecastConditionsDesc = prometheus.NewDesc(
		"weather_forecast",
		"Weather condition forecast",
		[]string{"location", "condition", "future_hours"},
		nil,
	)

	metricWeatherCurrentConditionsDesc = prometheus.NewDesc(
		"weather_current",
		"Weather condition current",
		[]string{"location", "condition"},
		nil,
	)

	metricPollutionCurrentDesc = prometheus.NewDesc(
		"pollution_current_aqi",
		"Current Air Pollution (AQI)",
		[]string{"location"},
		nil,
	)

	metricWeatherEpochDesc = prometheus.NewDesc(
		"weather_epoch",
		"Weather event: (sunrise|sunset|moonrise|moonset)",
		[]string{"location", "event"},
		nil,
	)

	metricWeatherSummaryDesc = prometheus.NewDesc(
		"weather_summary",
		"Weather description",
		[]string{"location", "main", "description"},
		nil,
	)
)

func (o *OWM) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricWeatherForecastConditionsDesc
	ch <- metricWeatherCurrentConditionsDesc
	ch <- metricWeatherEpochDesc
	ch <- metricPollutionCurrentDesc
}

func (o *OWM) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx, span := o.tracer.Start(ctx, "Collect")
	defer span.End()

	_ = level.Debug(o.logger).Log("msg", "collecting openweathermap data",
		"traceID", trace.SpanContextFromContext(ctx).TraceID().String(),
	)

	for _, location := range o.cfg.Locations {
		o.collectPollution(ctx, ch, location)
		o.collectOne(ctx, ch, location)
	}
}

func (o *OWM) collectPollution(ctx context.Context, ch chan<- prometheus.Metric, location Location) {
	_, span := o.tracer.Start(ctx, "collectPollution")
	defer span.End()

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	pollution, err := owm.NewPollution(o.cfg.APIKey, owm.WithHttpClient(o.client))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		_ = level.Error(o.logger).Log("msg", "failed to get new pollution data", "err", err)
	}

	params := &owm.PollutionParameters{
		Location: *coord,
		Datetime: "current",
	}

	if err := pollution.PollutionByParams(params); err != nil {
		span.SetStatus(codes.Error, err.Error())
		_ = level.Error(o.logger).Log("msg", "failed to update pollution data", "err", err)
	}

	for _, p := range pollution.List {
		ch <- prometheus.MustNewConstMetric(
			metricPollutionCurrentDesc,
			prometheus.GaugeValue,
			p.Main.Aqi,
			location.Name,
		)
	}
}

func (o *OWM) collectOne(ctx context.Context, ch chan<- prometheus.Metric, location Location) {
	ctx, span := o.tracer.Start(ctx, "collectOne")
	defer span.End()

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	// Possibility to exclude information. For example exclude daily information []string{ExcludeDaily}
	w, err := owm.NewOneCall("C", "EN", o.cfg.APIKey, []string{})
	if err != nil {
		level.Error(o.logger).Log("msg", "onecall failed", "err", err)
		return
	}

	err = w.OneCallByCoordinates(coord)
	if err != nil {
		level.Error(o.logger).Log("msg", "onecall coordinates failed", "err", err)
	}

	// Sunrise and sunset
	epochs := map[string]float64{
		"sunrise": float64(w.Current.Sunrise),
		"sunset":  float64(w.Current.Sunset),
	}

	for epoch, value := range epochs {
		ch <- prometheus.MustNewConstMetric(
			metricWeatherEpochDesc,
			prometheus.CounterValue,
			value,
			location.Name,
			epoch,
		)
	}

	// Current conditions
	currentConditions := map[string]float64{
		"clouds":      float64(w.Current.Clouds),
		"dew_point":   w.Current.DewPoint,
		"feels_like":  w.Current.FeelsLike,
		"humidity":    float64(w.Current.Humidity),
		"pressure":    float64(w.Current.Pressure),
		"rain_1h":     w.Current.Rain.OneH,
		"rain_3h":     w.Current.Rain.ThreeH,
		"snow_1h":     w.Current.Snow.OneH,
		"snow_3h":     w.Current.Snow.ThreeH,
		"temp":        w.Current.Temp,
		"uvi":         float64(w.Current.UVI),
		"visibility":  float64(w.Current.Visibility),
		"wind_degree": w.Current.WindDeg,
		"wind_gust":   w.Current.WindGust,
		"wind_speed":  w.Current.WindSpeed,
	}

	for condition, value := range currentConditions {
		if w.Current.Dt > 0 {
			ch <- prometheus.MustNewConstMetric(
				metricWeatherCurrentConditionsDesc,
				prometheus.GaugeValue,
				value,
				location.Name,
				condition,
			)
		}
	}

	for _, weather := range w.Current.Weather {
		if w.Current.Dt > 0 {
			o.weatherSummary(ctx, ch, location, weather)
		}
	}

	for _, hour := range w.Hourly {
		hourlyConditions := map[string]float64{
			"clouds":      float64(hour.Clouds),
			"dew_point":   hour.DewPoint,
			"feels_like":  hour.FeelsLike,
			"humidity":    float64(hour.Humidity),
			"pressure":    float64(hour.Pressure),
			"rain_1h":     hour.Rain.OneH,
			"rain_3h":     hour.Rain.ThreeH,
			"snow_1h":     hour.Snow.OneH,
			"snow_3h":     hour.Snow.ThreeH,
			"temp":        hour.Temp,
			"uvi":         float64(hour.UVI),
			"visibility":  float64(hour.Visibility),
			"wind_degree": hour.WindDeg,
			"wind_gust":   hour.WindGust,
			"wind_speed":  hour.WindSpeed,
		}

		if hour.Dt > 0 {

			i, err := strconv.ParseInt(strconv.Itoa(hour.Dt), 10, 64)
			if err != nil {
				_ = level.Error(o.logger).Log("failed to parse", "int", fmt.Sprintf("%d", hour.Dt))
				continue
			}

			tm := time.Until(time.Unix(i, 0)).Round(1 * time.Hour).Hours()

			for condition, value := range hourlyConditions {
				ch <- prometheus.MustNewConstMetric(
					metricWeatherForecastConditionsDesc,
					prometheus.GaugeValue,
					value,
					location.Name,
					condition,
					fmt.Sprintf("%dh", int(tm)),
				)
			}
		}

		// for _, weather := range hour.Weather {
		// 	if hour.Dt > 0 {
		// 		o.weatherSummary(ctx, ch, location, weather)
		// 	}
		// }

	}

}

func (o *OWM) weatherSummary(ctx context.Context, ch chan<- prometheus.Metric, location Location, summary owm.Weather) {
	ch <- prometheus.MustNewConstMetric(
		metricWeatherSummaryDesc,
		prometheus.CounterValue,
		1,
		location.Name,
		summary.Main,
		summary.Description,
	)
}
