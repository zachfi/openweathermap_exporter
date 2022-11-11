package owm

import (
	"context"
	"strconv"
	"time"

	owm "github.com/briandowns/openweathermap"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/otel/trace"
)

var (
	metricWeatherForecastConditionsDesc = prometheus.NewDesc(
		"weather_forecast",
		"Weather condition forecast",
		[]string{"location", "condition", "dt", "interval"},
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
		[]string{"location", "event", "dt"},
		nil,
	)

	metricWeatherSummaryDesc = prometheus.NewDesc(
		"weather_summary",
		"Weather description",
		[]string{"location", "main", "description", "dt"},
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

	ctx, span := o.tracer.Start(ctx, "owmCollect")
	defer span.End()

	_ = level.Debug(o.logger).Log("msg", "collecting openweathermap data",
		"traceID", trace.SpanContextFromContext(ctx).TraceID().String(),
	)

	for _, location := range o.cfg.Locations {
		// o.collectPollution(ctx, ch, location)
		o.collectOne(ctx, ch, location)
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
				metricWeatherForecastConditionsDesc,
				prometheus.GaugeValue,
				value,
				location.Name,
				condition,
				inHours,
			)
		}
	}
}

// func (o *OWM) collectPollution(ctx context.Context, ch chan<- prometheus.Metric, location Location) {
// 	ctx, span := o.tracer.Start(ctx, "collectPollution")
// 	defer span.End()
//
// 	coord := &owm.Coordinates{
// 		Longitude: location.Longitude,
// 		Latitude:  location.Latitude,
// 	}
//
// 	pollution, err := owm.NewPollution(o.cfg.APIKey, owm.WithHttpClient(o.client))
// 	if err != nil {
// 		span.SetStatus(codes.Error, err.Error())
// 		_ = level.Error(o.logger).Log("msg", "failed to get new pollution data", "err", err)
// 	}
//
// 	params := &owm.PollutionParameters{
// 		Location: *coord,
// 		Datetime: "current",
// 	}
//
// 	if err := pollution.PollutionByParams(params); err != nil {
// 		span.SetStatus(codes.Error, err.Error())
// 		_ = level.Error(o.logger).Log("msg", "failed to update pollution data", "err", err)
// 	}
//
// 	for _, p := range pollution.List {
// 		ch <- prometheus.MustNewConstMetric(
// 			metricPollutionCurrentDesc,
// 			prometheus.GaugeValue,
// 			p.Main.Aqi,
// 			location.Name,
// 		)
// 	}
// }

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
			strconv.Itoa(w.Current.Dt),
		)
	}

	// Current conditions
	conditions := map[string]float64{
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

	for condition, value := range conditions {

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
		o.weatherSummary(ctx, ch, location, w.Current.Dt, weather)
	}

	// Minutely forecast
	for _, value := range w.Minutely {
		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastConditionsDesc,
			prometheus.GaugeValue,
			value.Precipitation,
			location.Name,
			"precipitation",
			strconv.Itoa(value.Dt),
			"minute",
		)
	}

	// Hourly forecast
	for _, hour := range w.Hourly {
		conditions := map[string]float64{
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

		for condition, value := range conditions {
			ch <- prometheus.MustNewConstMetric(
				metricWeatherForecastConditionsDesc,
				prometheus.GaugeValue,
				value,
				location.Name,
				condition,
				strconv.Itoa(hour.Dt),
				"hour",
			)
		}

		for _, weather := range hour.Weather {
			o.weatherSummary(ctx, ch, location, hour.Dt, weather)
		}

	}

	// Daily forecast
	for _, day := range w.Daily {
		conditions := map[string]float64{
			"clouds":                    float64(day.Clouds),
			"dew_point":                 day.DewPoint,
			"humidity":                  float64(day.Humidity),
			"pressure":                  float64(day.Pressure),
			"rain":                      day.Rain,
			"snow":                      day.Snow,
			"temp_min":                  day.Temp.Min,
			"temp_max":                  day.Temp.Max,
			"uvi":                       float64(day.UVI),
			"wind_degree":               day.WindDeg,
			"wind_gust":                 day.WindGust,
			"wind_speed":                day.WindSpeed,
			"precipitation_probability": day.Pop,
		}

		for condition, value := range conditions {
			ch <- prometheus.MustNewConstMetric(
				metricWeatherForecastConditionsDesc,
				prometheus.GaugeValue,
				value,
				location.Name,
				condition,
				strconv.Itoa(day.Dt),
				"day",
			)
		}

		// Moonrise and moonset
		epochs := map[string]float64{
			"sunrise":   float64(day.Sunrise),
			"sunset":    float64(day.Sunset),
			"moonrise":  float64(day.Moonrise),
			"moonset":   float64(day.Moonset),
			"moonphase": float64(day.MoonPhase),
		}

		for epoch, value := range epochs {
			ch <- prometheus.MustNewConstMetric(
				metricWeatherEpochDesc,
				prometheus.CounterValue,
				value,
				location.Name,
				epoch,
				strconv.Itoa(day.Dt),
			)
		}
	}
}

func (o *OWM) weatherSummary(ctx context.Context, ch chan<- prometheus.Metric, location Location, dt int, summary owm.Weather) {
	ch <- prometheus.MustNewConstMetric(
		metricWeatherSummaryDesc,
		prometheus.CounterValue,
		1,
		location.Name,
		summary.Main,
		summary.Description,
		strconv.Itoa(dt),
	)
}
