package exporter

import (
	"strconv"

	owm "github.com/briandowns/openweathermap"
	"github.com/prometheus/client_golang/prometheus"
)

type Forecast struct {
	High float32
	Low  float32
}

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

	moonRiseTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_moonrise_time",
		Help: "Time of Moon Rise",
	}, nil)

	moonSetTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_moonrise_set",
		Help: "Time of Moon Set",
	}, nil)

	sunRiseTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_sunrise_time",
		Help: "Time of Sun Rise",
	}, nil)

	sunSetTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "owm_sunrise_set",
		Help: "Time of Sun Set",
	}, nil)
)

func init() {
	prometheus.MustRegister(
		currentTemp,
		forecastHighTemp,
		forecastLowTemp,
		moonRiseTime,
		moonSetTime,
		sunRiseTime,
		sunSetTime,
	)
}

// func forecastWatch(apiKey string) error {
// 	log.Warnf("FUCK: %s", "FUCK")
// 	c := gowu.NewClient(apiKey)
// 	fore, err := c.GetForecast("portland", "or")
// 	if err != nil {
// 		log.Warnf("ERR: %s", "err")
// 		return err
// 	}
//
// 	log.Warnf("Result: %+v", fore)
//
// 	for i, day := range fore.Simpleforecast.Forecastday {
// 		dayString := strconv.Itoa(i)
//
// 		highTemp, err := strconv.ParseFloat(day.High.Celsius, 32)
// 		if err != nil {
// 			log.Error(err)
// 		}
//
// 		lowTemp, err := strconv.ParseFloat(day.Low.Celsius, 32)
// 		if err != nil {
// 			log.Error(err)
// 		}
//
// 		forecastHighTemp.With(prometheus.Labels{"day": dayString}).Set(highTemp)
// 		forecastLowTemp.With(prometheus.Labels{"day": dayString}).Set(lowTemp)
// 	}
//
// 	return nil
// }
//
// func astroWatch(apiKey string) error {
// 	c := gowu.NewClient(apiKey)
// 	moonPhase, sunPhase, err := c.GetAstronomy("portland", "or")
// 	if err != nil {
// 		return err
// 	}
//
// 	log.Debugf("Result: %+v", moonPhase)
//
// 	moonRiseHourMin, err := strconv.ParseFloat(
// 		fmt.Sprintf("%s.%s", moonPhase.MoonRise.Hour, moonPhase.MoonRise.Minute), 32)
// 	moonSetHourMin, err := strconv.ParseFloat(
// 		fmt.Sprintf("%s.%s", moonPhase.MoonSet.Hour, moonPhase.MoonSet.Minute), 32)
//
// 	sunRiseHourMin, err := strconv.ParseFloat(
// 		fmt.Sprintf("%s.%s", sunPhase.SunRise.Hour, sunPhase.SunRise.Minute), 32)
// 	sunSetHourMin, err := strconv.ParseFloat(
// 		fmt.Sprintf("%s.%s", sunPhase.SunSet.Hour, sunPhase.SunSet.Minute), 32)
//
// 	moonRiseTime.With(prometheus.Labels{}).Set(moonRiseHourMin)
// 	moonSetTime.With(prometheus.Labels{}).Set(moonSetHourMin)
//
// 	sunRiseTime.With(prometheus.Labels{}).Set(sunRiseHourMin)
// 	sunSetTime.With(prometheus.Labels{}).Set(sunSetHourMin)
//
// 	return nil
// }

func forecast(apiKey string) error {
	w, err := owm.NewForecast("5", "C", "EN", apiKey) // valid options for first parameter are "5" and "16"
	if err != nil {
		return err
	}

	coord := &owm.Coordinates{
		Longitude: -122.588975,
		Latitude:  45.554587,
	}

	err = w.DailyByCoordinates(
		coord,
		5,
	)
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
		// log.Infof("P: %+v", p)
		inHours := strconv.Itoa(i * 3)

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

	// log.Infof("%+v", c)
	currentTemp.With(prometheus.Labels{}).Set(c.Main.Temp)

	sunRiseTime.With(prometheus.Labels{}).Set(float64(c.Sys.Sunrise))
	sunSetTime.With(prometheus.Labels{}).Set(float64(c.Sys.Sunset))

	return nil
}

func ScrapeMetrics(apiKey string) error {
	err := forecast(apiKey)
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
