package owm

import (
	"strconv"

	owm "github.com/briandowns/openweathermap"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
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

	metricWeatherEpoch = prometheus.NewDesc(
		"weather_epoch",
		"Weather event, (sunrise|sunset)",
		[]string{"location", "event"},
		nil,
	)
)

func (o *OWM) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricWeatherForecastDesc
	ch <- metricWeatherCurrentDesc
	ch <- metricWeatherEpoch
}

func (o *OWM) Collect(ch chan<- prometheus.Metric) {

	for _, location := range o.cfg.Locations {
		//
		// handle wind gust - Probably needs a PR upstream
		// handle UTC DT
		// handle feels_like - needs upstream PR

		o.collectForecast(ch, location)
		o.collectCurrent(ch, location)
	}

}

func (o *OWM) collectForecast(ch chan<- prometheus.Metric, location Location) {

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	w, err := owm.NewForecast("5", "C", "EN", o.cfg.APIKey) // valid options for first parameter are "5" and "16"
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

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Main.Temp,
			location.Name,
			"temp",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Main.TempMax,
			location.Name,
			"temp_max",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Main.TempMin,
			location.Name,
			"temp_min",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Main.GrndLevel,
			location.Name,
			"pressure",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			float64(p.Main.Humidity),
			location.Name,
			"humidity",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Wind.Speed,
			location.Name,
			"wind_speed",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Wind.Deg,
			location.Name,
			"wind_degree",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Rain.ThreeH,
			location.Name,
			"rain",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			p.Snow.ThreeH,
			location.Name,
			"snow",
			inHours,
		)

		ch <- prometheus.MustNewConstMetric(
			metricWeatherForecastDesc,
			prometheus.GaugeValue,
			float64(p.Clouds.All),
			location.Name,
			"clouds",
			inHours,
		)

	}

}

func (o *OWM) collectCurrent(ch chan<- prometheus.Metric, location Location) {

	coord := &owm.Coordinates{
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}

	c, err := owm.NewCurrent("C", "EN", o.cfg.APIKey)
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
		metricWeatherEpoch,
		prometheus.CounterValue,
		float64(c.Sys.Sunrise),
		location.Name,
		"sunrise",
	)

	// sunset := time.Unix(int64(c.Sys.Sunset), 0)
	// _ = level.Info(o.logger).Log("msg", "current", "sunset", fmt.Sprintf("%+v", sunset.String()))

	ch <- prometheus.MustNewConstMetric(
		metricWeatherEpoch,
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
	// 	"current_wind_gust",
	// )

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Snow.ThreeH,
		location.Name,
		"current_snow_threeh",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Rain.ThreeH,
		location.Name,
		"current_rain_threeh",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		float64(c.Clouds.All),
		location.Name,
		"current_cloud_cover",
	)
}
