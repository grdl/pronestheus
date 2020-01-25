package nest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/giantswarm/micrologger"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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

// Config provides the necessary configuration for creating a Collector
type Config struct {
	Logger  micrologger.Logger
	Timeout int

	ApiURL   string
	ApiToken string
}

// Collector implements the Collector interface, collecting weather data from OpenWeatherMap APi
type Collector struct {
	logger  micrologger.Logger
	timeout time.Duration

	apiURL   string
	apiToken string

	up       *prometheus.Desc
	temp     *prometheus.Desc
	target   *prometheus.Desc
	humidity *prometheus.Desc
	heating  *prometheus.Desc
	leaf     *prometheus.Desc
	sunlight *prometheus.Desc

	//TODO:
	// errors?
}

// New creates a Collector with given Config
func New(config Config) (*Collector, error) {
	if config.Logger == nil {
		return nil, errors.New("Logger must not be empty")
	}
	if config.Timeout <= 0 {
		return nil, errors.New("Timeout must not be empty")
	}
	if config.ApiURL == "" {
		return nil, errors.New("Nest Api URL config must not be empty")
	}
	if config.ApiToken == "" {
		return nil, errors.New("Nest Api Token config must not be empty")
	}

	var nestLabels = []string{"id", "name"}

	collector := &Collector{
		logger:   config.Logger,
		timeout:  time.Duration(config.Timeout) * time.Millisecond,
		apiURL:   config.ApiURL,
		apiToken: config.ApiToken,
		up:       prometheus.NewDesc("nest_up", "Was talking to Nest API successful.", nil, nil),
		temp:     prometheus.NewDesc("nest_current_temp", "Current ambient temperature.", nestLabels, nil),
		target:   prometheus.NewDesc("nest_target_temp", "Current target temperature.", nestLabels, nil),
		humidity: prometheus.NewDesc("nest_humidity", "Current inside humidity.", nestLabels, nil),
		heating:  prometheus.NewDesc("nest_heating", "Is thermostat heating.", nestLabels, nil),
		leaf:     prometheus.NewDesc("nest_leaf", "Is thermostat set to energy-saving temperature.", nestLabels, nil),
		sunlight: prometheus.NewDesc("nest_sunlight", "Is thermostat in direct sunlight.", nestLabels, nil),
	}

	return collector, nil
}

// Implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.temp
	ch <- c.target
	ch <- c.humidity
	ch <- c.heating
	ch <- c.leaf
	ch <- c.sunlight
}

// Implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "Scraping Nest API")

	thermostats, err := c.getNestReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		c.logger.Log("level", "error", "message", "could not get nest readings", "stack", errors.WithStack(err))
		return
	}

	c.logger.Log("level", "debug", "message", "Successfully scraped Nest API")

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)

	for _, therm := range thermostats {
		labels := []string{therm.Id, strings.Replace(therm.Name, " ", "-", -1)}

		ch <- prometheus.MustNewConstMetric(c.temp, prometheus.GaugeValue, therm.Temperature, labels...)
		ch <- prometheus.MustNewConstMetric(c.target, prometheus.GaugeValue, therm.Target, labels...)
		ch <- prometheus.MustNewConstMetric(c.humidity, prometheus.GaugeValue, therm.Humidity, labels...)
		ch <- prometheus.MustNewConstMetric(c.heating, prometheus.GaugeValue, b2f(therm.HVACState == "heating"), labels...)
		ch <- prometheus.MustNewConstMetric(c.leaf, prometheus.GaugeValue, b2f(therm.Leaf), labels...)
		ch <- prometheus.MustNewConstMetric(c.sunlight, prometheus.GaugeValue, b2f(therm.Sunlight), labels...)

	}
}

func (c *Collector) getNestReadings() (thermostats []Thermostat, err error) {
	req, _ := http.NewRequest("GET", c.apiURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	// Nest API needs a custom http client to be able to handle redirects
	// See https://developers.nest.com/guides/api/how-to-handle-redirects
	client := http.Client{
		Timeout: c.timeout,
		CheckRedirect: func(redirReq *http.Request, via []*http.Request) error {
			redirReq.Header = req.Header

			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	res, err := client.Do(req)
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

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
