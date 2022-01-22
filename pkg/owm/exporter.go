package owm

import (
	"fmt"
	"time"

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

	metricWeatherForecastTempDesc = prometheus.NewDesc(
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

	// fore := w.ForecastWeatherJson.(*owm.Forecast5WeatherData)

	// for i, p := range fore.List {
	// inHours := strconv.Itoa(i * 3)

	// ch <- prometheus.MustNewConstMetric(
	// 	metricWeatherForecastDesc,
	// 	prometheus.GaugeValue,
	// 	p.Rain.ThreeH,
	// 	location.Name,
	// 	"rain",
	// 	inHours,
	// )

	// ch <- prometheus.MustNewConstMetric(
	// 	metricWeatherForecastDesc,
	// 	prometheus.GaugeValue,
	// 	p.Snow.ThreeH,
	// 	location.Name,
	// 	"snow",
	// 	inHours,
	// )

	// ch <- prometheus.MustNewConstMetric(
	// 	metricWeatherForecastDesc,
	// 	prometheus.GaugeValue,
	// 	float64(p.Clouds.All),
	// 	location.Name,
	// 	"clouds",
	// 	inHours,
	// )

	// ch <- prometheus.MustNewConstMetric(
	// 	metricWeatherForecastDesc,
	// 	prometheus.GaugeValue,
	// 	float64(p.Clouds.All),
	// 	location.Name,
	// 	"clouds",
	// 	inHours,
	// )

	// }

	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.MaxLocalTracesPerUser), MetricMaxLocalTracesPerUser, tenant)
	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.MaxGlobalTracesPerUser), MetricMaxGlobalTracesPerUser, tenant)
	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.MaxBytesPerTrace), MetricMaxBytesPerTrace, tenant)
	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.MaxSearchBytesPerTrace), MetricMaxSearchBytesPerTrace, tenant)
	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.IngestionRateLimitBytes), MetricIngestionRateLimitBytes, tenant)
	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.IngestionBurstSizeBytes), MetricIngestionBurstSizeBytes, tenant)
	// ch <- prometheus.MustNewConstMetric(metricOverridesLimitsDesc, prometheus.GaugeValue, float64(limits.BlockRetention), MetricBlockRetention, tenant)

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

	sunrise := time.Unix(int64(c.Sys.Sunrise), 0)
	_ = level.Info(o.logger).Log("msg", "current", "sunrise", fmt.Sprintf("%+v", sunrise.String()))

	ch <- prometheus.MustNewConstMetric(
		metricWeatherEpoch,
		prometheus.CounterValue,
		float64(c.Sys.Sunrise),
		location.Name,
		"sunrise",
	)

	sunset := time.Unix(int64(c.Sys.Sunset), 0)
	_ = level.Info(o.logger).Log("msg", "current", "sunset", fmt.Sprintf("%+v", sunset.String()))

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
		"current_temp",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.TempMax,
		location.Name,
		"current_temp_max",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.TempMin,
		location.Name,
		"current_temp_min",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Main.GrndLevel,
		location.Name,
		"current_pressure",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		float64(c.Main.Humidity),
		location.Name,
		"current_humidity",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Wind.Speed,
		location.Name,
		"current_wind_speed",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		c.Wind.Deg,
		location.Name,
		"current_wind_degree",
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
		float64(c.Snow.ThreeH),
		location.Name,
		"current_snow_threeh",
	)

	ch <- prometheus.MustNewConstMetric(
		metricWeatherCurrentDesc,
		prometheus.CounterValue,
		float64(c.Rain.ThreeH),
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

	// currentTemp.With(prometheus.Labels{"location": locationName}).Set(c.Main.Temp)

	// sunRiseTime.With(prometheus.Labels{"location": locationName}).Set(float64(c.Sys.Sunrise))
	// sunSetTime.With(prometheus.Labels{"location": locationName}).Set(float64(c.Sys.Sunset))

	// windSpeed.With(prometheus.Labels{"location": locationName}).Set(float64(c.Wind.Speed))
	// windDegrees.With(prometheus.Labels{"location": locationName}).Set(float64(c.Wind.Deg))

	// snow.With(prometheus.Labels{"location": locationName}).Set(float64(c.Snow.ThreeH))
	// rain.With(prometheus.Labels{"location": locationName}).Set(float64(c.Rain.ThreeH))
	// clouds.With(prometheus.Labels{"location": locationName}).Set(float64(c.Clouds.All))
}
