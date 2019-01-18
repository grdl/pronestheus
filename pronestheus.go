package main

import (
	"encoding/json"

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

//Weather stores weather data received from Open Weather Map API
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
}

// Implements prometheus.Collector
func (c NestCollector) Collect(ch chan<- prometheus.Metric) {
	thermostats, err := getNestReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(nestUp, prometheus.GaugeValue, 0)
		log.Error(err)
		return
	}

	ch <- prometheus.MustNewConstMetric(nestUp, prometheus.GaugeValue, 1)

	for _, therm := range thermostats {
		labels := []string{therm.Id, therm.Name}

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

	log.Info(thermostats)
}

func getNestReadings() (thermostats []Thermostat, err error) {

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

	// TODO: should return error if response is not 2xx

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Reading Nest API response failed")
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

func main() {
	kingpin.Parse()

	c := NestCollector{}
	prometheus.MustRegister(c)

	log.With("addr", *listenAddress).Info("Started listening")

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
