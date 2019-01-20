package main

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/alecthomas/kingpin.v2"

	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	listenAddress = kingpin.Flag("listen-addr", "The address to listen on").Default(":2112").String()
	nestApiURL    = kingpin.Flag("nest-api-url", "The Nest API URL").Default("https://developer-api.nest.com/devices/thermostats").String()
	nestApiToken  = kingpin.Flag("nest-api-token", "The authorization token for Nest API").Required().String()

	weatherApiURL        = kingpin.Flag("weather-api-url", "The OpenWeatherMap URL").Default("http://api.openweathermap.org/data/2.5/weather").String()
	weatherApiToken      = kingpin.Flag("weather-api-token", "The authorization token for OpenWeatherMap API").Default("").String()
	weatherApiLocationId = kingpin.Flag("weather-api-location-id", "The location ID for OpenWeatherMap API. Defaults to Amsterdam").Default("2759794").String()
)

var nestLabels = []string{"id", "name"}

var (
	nestUp       = prometheus.NewDesc("nest_up", "Was talking to Nest API successful.", nil, nil)
	nestTemp     = prometheus.NewDesc("nest_current_temp", "Current ambient temperature.", nestLabels, nil)
	nestTarget   = prometheus.NewDesc("nest_target_temp", "Current target temperature.", nestLabels, nil)
	nestHumidity = prometheus.NewDesc("nest_humidity", "Current inside humidity.", nestLabels, nil)
	nestHeating  = prometheus.NewDesc("nest_heating", "Is thermostat heating.", nestLabels, nil)
	nestLeaf     = prometheus.NewDesc("nest_leaf", "Is thermostat set to energy-saving temperature.", nestLabels, nil)
	nestSunlight = prometheus.NewDesc("nest_sunlight", "Is thermostat in direct sunlight.", nestLabels, nil)
)

var (
	weatherUp       = prometheus.NewDesc("nest_weather_up", "Was talking to OpenWeatherMap API successful.", nil, nil)
	weatherTemp     = prometheus.NewDesc("nest_weather_temp", "Current outside temperature.", nil, nil)
	weatherHumidity = prometheus.NewDesc("nest_weather_humidity", "Current outside humidity", nil, nil)
	weatherPressure = prometheus.NewDesc("nest_weather_pressure", "Current outside pressure", nil, nil)
)

// Thermostat stores thermostat readings received from Nest API
type Thermostat struct {
	Id          string  `json:"device_id"`
	Name        string  `json:"name"`
	Temperature float64 `json:"ambient_temperature_c"`
	Target      float64 `json:"target_temperature_c"`
	Humidity    float64 `json:"humidity"`
	HVACState   string  `json:"hvac_state"`
	Leaf        bool    `json:"has_leaf"`
	Sunlight    bool    `json:"sunlight_correction_active"`
}

//Weather stores weather data received from OpenWeatherMap API
type Weather struct {
	Temperature float64 `json:"temp"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}

type NestCollector struct {
}

// Implements prometheus.Collector
func (c NestCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nestUp
	ch <- nestTemp
	ch <- nestTarget
	ch <- nestHumidity
	ch <- nestHeating
	ch <- nestLeaf
	ch <- nestSunlight
}

// Implements prometheus.Collector
func (c NestCollector) Collect(ch chan<- prometheus.Metric) {
	log.Info("Scraping Nest API")

	thermostats, err := c.getNestReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(nestUp, prometheus.GaugeValue, 0)
		log.Error(err)
		return
	}

	log.Info("Successfully scraped Nest API")

	ch <- prometheus.MustNewConstMetric(nestUp, prometheus.GaugeValue, 1)

	for _, therm := range thermostats {
		labels := []string{therm.Id, strings.Replace(therm.Name, " ", "-", -1)}

		ch <- prometheus.MustNewConstMetric(nestTemp, prometheus.GaugeValue, therm.Temperature, labels...)
		ch <- prometheus.MustNewConstMetric(nestTarget, prometheus.GaugeValue, therm.Target, labels...)
		ch <- prometheus.MustNewConstMetric(nestHumidity, prometheus.GaugeValue, therm.Humidity, labels...)

		if therm.HVACState == "heating" {
			ch <- prometheus.MustNewConstMetric(nestHeating, prometheus.GaugeValue, 1, labels...)
		} else {
			ch <- prometheus.MustNewConstMetric(nestHeating, prometheus.GaugeValue, 0, labels...)
		}

		if therm.Leaf {
			ch <- prometheus.MustNewConstMetric(nestLeaf, prometheus.GaugeValue, 1, labels...)
		} else {
			ch <- prometheus.MustNewConstMetric(nestLeaf, prometheus.GaugeValue, 0, labels...)
		}

		if therm.Sunlight {
			ch <- prometheus.MustNewConstMetric(nestSunlight, prometheus.GaugeValue, 1, labels...)
		} else {
			ch <- prometheus.MustNewConstMetric(nestSunlight, prometheus.GaugeValue, 0, labels...)
		}
	}
}

func (c NestCollector) getNestReadings() (thermostats []Thermostat, err error) {
	req, _ := http.NewRequest("GET", *nestApiURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *nestApiToken))

	// Nest API needs a custom http client to be able to handle redirects
	// See https://developers.nest.com/guides/api/how-to-handle-redirects
	customClient := http.Client{
		CheckRedirect: func(redirReq *http.Request, via []*http.Request) error {
			redirReq.Header = req.Header

			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	res, err := customClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Calling Nest API failed")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Reading Nest API response failed")
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Nest API responded with %d code: %s", res.StatusCode, body))
	}

	var data map[string]Thermostat
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshalling Nest API response failed")
	}

	for _, thermostat := range data {
		thermostats = append(thermostats, thermostat)
	}

	return thermostats, nil
}

type WeatherCollector struct {
}

// Implements prometheus.Collector
func (c WeatherCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- weatherUp
	ch <- weatherTemp
	ch <- weatherHumidity
	ch <- weatherPressure
}

// Implements prometheus.Collector
func (c WeatherCollector) Collect(ch chan<- prometheus.Metric) {
	log.Info("Scraping OpenWeatherMap API")

	weather, err := c.getWeatherReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(weatherUp, prometheus.GaugeValue, 0)
		log.Error(err)
		return
	}

	log.Info("Successfully scraped OpenWeatherMap API")

	ch <- prometheus.MustNewConstMetric(weatherUp, prometheus.GaugeValue, 1)

	ch <- prometheus.MustNewConstMetric(weatherTemp, prometheus.GaugeValue, weather.Temperature)
	ch <- prometheus.MustNewConstMetric(weatherHumidity, prometheus.GaugeValue, weather.Humidity)
	ch <- prometheus.MustNewConstMetric(weatherPressure, prometheus.GaugeValue, weather.Pressure)
}

func (c WeatherCollector) getWeatherReadings() (weather Weather, err error) {
	url := fmt.Sprintf("%s?id=%s&appid=%s&units=metric", *weatherApiURL, *weatherApiLocationId, *weatherApiToken)

	res, err := http.Get(url)
	if err != nil {
		return weather, errors.Wrap(err, "Calling  API failed")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return weather, errors.Wrap(err, "Reading OpenWeatherMap API response failed")
	}

	if res.StatusCode != 200 {
		return weather, errors.New(fmt.Sprintf("OpenWeatherMap responded with %d code: %s", res.StatusCode, body))
	}

	var data map[string]json.RawMessage

	err = json.Unmarshal(body, &data)
	if err != nil {
		return weather, errors.Wrap(err, "Unmarshalling OpenWeatherMap API response failed")
	}

	err = json.Unmarshal(data["main"], &weather)
	if err != nil {
		return weather, errors.Wrap(err, "Unmarshalling OpenWeatherMap API response failed")
	}

	return weather, nil
}

func main() {
	kingpin.Parse()

	c := NestCollector{}
	prometheus.MustRegister(c)

	if *weatherApiToken != "" {
		w := WeatherCollector{}
		prometheus.MustRegister(w)
	}

	log.With("listening_addr", *listenAddress).Info("Started Pronestheus - Nest Thermostat Prometheus Exporter")

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
