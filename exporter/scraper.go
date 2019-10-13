package exporter

import (
	"strconv"

	owm "github.com/briandowns/openweathermap"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	currentTemp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_current_temperature",
		Help: "Temperature in Celcius",
	}, []string{})

	forecastHighTemp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_forecast_high_temperature",
		Help: "Temperature in Celcius",
	}, []string{"inhours"})

	forecastLowTemp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_forecast_low_temperature",
		Help: "Temperature in Celcius",
	}, []string{"inhours"})

	sunRiseTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_sunrise_time",
		Help: "Time of Sun Rise",
	}, nil)

	sunSetTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_sunset_time",
		Help: "Time of Sun Set",
	}, nil)

	windSpeed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_current_wind_speed",
		Help: "The current speed of the wind",
	}, nil)

	windDegrees = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_current_wind_degrees",
		Help: "The current degreese of the wind",
	}, nil)

	rain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_current_rain",
		Help: "The current rain",
	}, nil)

	snow = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_current_snow",
		Help: "The current snow",
	}, nil)

	clouds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_current_clouds",
		Help: "The current clouds",
	}, nil)
)

func init() {
	prometheus.MustRegister(
		currentTemp,
		forecastHighTemp,
		forecastLowTemp,
		sunRiseTime,
		sunSetTime,
		windSpeed,
		windDegrees,
		rain,
		snow,
		clouds,
	)
}

func forecast(apiKey string, coord *owm.Coordinates) error {
	w, err := owm.NewForecast("5", "C", "EN", apiKey) // valid options for first parameter are "5" and "16"
	if err != nil {
		return err
	}

	err = w.DailyByCoordinates(coord, 50)
	if err != nil {
		return err
	}

	fore := w.ForecastWeatherJson.(*owm.Forecast5WeatherData)

	// log.Infof("Type: %T", w.ForecastWeatherJson)
	// log.Infof("Result: %+v", w.ForecastWeatherJson.(*owm.Forecast5WeatherData))
	// log.Infof("Result: %T", w)
	// log.Infof("Result: %+v", w)
	// log.Infof("Entry count: %d", fore.Cnt)

	for i, p := range fore.List {
		inHours := strconv.Itoa(i * 3)
		// TODO parse time from forecast and calculate distance

		// log.Debugf("Weather: %+v", p.Weather)

		forecastHighTemp.With(prometheus.Labels{"inhours": inHours}).Set(p.Main.TempMax)
		forecastLowTemp.With(prometheus.Labels{"inhours": inHours}).Set(p.Main.TempMin)
	}

	c, err := owm.NewCurrent("C", "EN", apiKey)
	if err != nil {
		return err
	}

	err = c.CurrentByCoordinates(coord)
	if err != nil {
		return err
	}

	log.Debugf("Scrape for: %s", c.Name)

	// log.Infof("%+v", c)
	currentTemp.With(prometheus.Labels{}).Set(c.Main.Temp)

	sunRiseTime.With(prometheus.Labels{}).Set(float64(c.Sys.Sunrise))
	sunSetTime.With(prometheus.Labels{}).Set(float64(c.Sys.Sunset))

	windSpeed.With(prometheus.Labels{}).Set(float64(c.Wind.Speed))
	windDegrees.With(prometheus.Labels{}).Set(float64(c.Wind.Deg))

	snow.With(prometheus.Labels{}).Set(float64(c.Snow.ThreeH))
	rain.With(prometheus.Labels{}).Set(float64(c.Rain.ThreeH))
	clouds.With(prometheus.Labels{}).Set(float64(c.Clouds.All))

	return nil
}

func ScrapeMetrics(apiKey string, longitude, latitude float64) error {

	coord := &owm.Coordinates{
		Longitude: longitude,
		Latitude:  latitude,
	}

	err := forecast(apiKey, coord)
	if err != nil {
		return err
	}
	//
	// err = astroWatch(apiKey)
	// if err != nil {
	// 	return err
	// }

	return nil
}
