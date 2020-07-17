package nest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// Thermostat stores thermostat data received from Nest API.
type Thermostat struct {
	ID          string  `json:"device_id"`
	Name        string  `json:"name"`
	Temperature float64 `json:"ambient_temperature_c"`
	Target      float64 `json:"target_temperature_c"`
	Humidity    float64 `json:"humidity"`
	HVACState   string  `json:"hvac_state"`
	Leaf        bool    `json:"has_leaf"`
	Sunlight    bool    `json:"sunlight_correction_active"`
}

// Config provides the configuration necessary to create the Collector.
type Config struct {
	Logger   log.Logger
	Timeout  int
	Unit     string
	APIURL   string
	APIToken string
}

// Collector implements the Collector interface, collecting thermostats data from Nest API.
type Collector struct {
	client   *http.Client
	req      *http.Request
	logger   log.Logger
	up       *prometheus.Desc
	temp     *prometheus.Desc
	target   *prometheus.Desc
	humidity *prometheus.Desc
	heating  *prometheus.Desc
	leaf     *prometheus.Desc
	sunlight *prometheus.Desc
}

// New creates a Collector using the given Config.
func New(cfg Config) (*Collector, error) {
	req, err := http.NewRequest("GET", cfg.APIURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating Nest API request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIToken))

	// Nest API needs a custom http client to be able to add headers to redirects.
	// See https://developers.nest.com/guides/api/how-to-handle-redirects
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Millisecond,
		CheckRedirect: func(redirReq *http.Request, via []*http.Request) error {
			redirReq.Header = req.Header

			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	var nestLabels = []string{"id", "name"}
	collector := &Collector{
		client:   client,
		req:      req,
		logger:   cfg.Logger,
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

// Describe implements the prometheus.Describe interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.temp
	ch <- c.target
	ch <- c.humidity
	ch <- c.heating
	ch <- c.leaf
	ch <- c.sunlight
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	thermostats, err := c.getNestReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		c.logger.Log("level", "error", "message", "Failed collecting Nest data", "stack", errors.WithStack(err))
		return
	}

	c.logger.Log("level", "debug", "message", "Successfully collected Nest data")

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)

	for _, therm := range thermostats {
		labels := []string{therm.ID, strings.Replace(therm.Name, " ", "-", -1)}

		ch <- prometheus.MustNewConstMetric(c.temp, prometheus.GaugeValue, therm.Temperature, labels...)
		ch <- prometheus.MustNewConstMetric(c.target, prometheus.GaugeValue, therm.Target, labels...)
		ch <- prometheus.MustNewConstMetric(c.humidity, prometheus.GaugeValue, therm.Humidity, labels...)
		ch <- prometheus.MustNewConstMetric(c.heating, prometheus.GaugeValue, b2f(therm.HVACState == "heating"), labels...)
		ch <- prometheus.MustNewConstMetric(c.leaf, prometheus.GaugeValue, b2f(therm.Leaf), labels...)
		ch <- prometheus.MustNewConstMetric(c.sunlight, prometheus.GaugeValue, b2f(therm.Sunlight), labels...)

	}
}

func (c *Collector) getNestReadings() (thermostats []Thermostat, err error) {
	res, err := c.client.Do(c.req)
	if err != nil {
		return nil, errors.Wrap(err, "calling Nest API failed")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading Nest API response failed")
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("nest API responded with %d code: %s", res.StatusCode, body))
	}

	var data map[string]Thermostat
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling Nest API response failed")
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
