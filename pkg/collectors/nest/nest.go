package nest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	celsius    string = "celsius"
	fahrenheit string = "fahrenheit"
)

var (
	errNon200Response        = errors.New("nest API responded with non-200 code")
	errFailedCreatingRequest = errors.New("failed creating Nest API request")
	errFailedParsingURL      = errors.New("failed parsing OpenWeatherMap API URL")
	errInvalidTempUnit       = errors.New("invalid temperature unit; valid values: [celsius, fahrenheit]")
	errFailedUnmarshalling   = errors.New("failed unmarshalling Nest API response body")
	errFailedRequest         = errors.New("failed Nest API request")
	errFailedReadingBody     = errors.New("failed reading Nest API response body")
	errReachedMaxRedirects   = errors.New("reached max redirects")
)

// Thermostat stores thermostat data received from Nest API.
type Thermostat struct {
	ID           string  `json:"device_id"`
	Name         string  `json:"name"`
	TemperatureC float64 `json:"ambient_temperature_c"`
	TemperatureF float64 `json:"ambient_temperature_f"`
	TargetC      float64 `json:"target_temperature_c"`
	TargetF      float64 `json:"target_temperature_f"`
	Humidity     float64 `json:"humidity"`
	HVACState    string  `json:"hvac_state"`
	Leaf         bool    `json:"has_leaf"`
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
	client  *http.Client
	req     *http.Request
	logger  log.Logger
	metrics *Metrics
	unit    string
}

// Metrics contains the metrics collected by the Collector.
type Metrics struct {
	up       *prometheus.Desc
	temp     *prometheus.Desc
	target   *prometheus.Desc
	humidity *prometheus.Desc
	heating  *prometheus.Desc
	leaf     *prometheus.Desc
}

// New creates a Collector using the given Config.
func New(cfg Config) (*Collector, error) {
	if _, err := url.ParseRequestURI(cfg.APIURL); err != nil {
		return nil, errors.Wrap(errFailedParsingURL, err.Error())
	}

	req, err := http.NewRequest(http.MethodGet, cfg.APIURL, nil)
	if err != nil {
		return nil, errors.Wrap(errFailedCreatingRequest, err.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIToken))

	// Nest API needs a custom http client to be able to pass the auth header to redirect destination.
	// See https://developers.nest.com/guides/api/how-to-handle-redirects
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Millisecond,
		CheckRedirect: func(redirReq *http.Request, via []*http.Request) error {
			redirReq.Header = req.Header

			if len(via) >= 10 {
				return errReachedMaxRedirects
			}
			return nil
		},
	}

	collector := &Collector{
		client:  client,
		req:     req,
		logger:  cfg.Logger,
		metrics: buildMetrics(cfg.Unit),
		unit:    cfg.Unit,
	}

	return collector, nil
}

func buildMetrics(unit string) *Metrics {
	if unit == "" {
		unit = celsius
	}

	var nestLabels = []string{"id", "name"}
	return &Metrics{
		up:       prometheus.NewDesc(strings.Join([]string{"nest", "up"}, "_"), "Was talking to Nest API successful.", nil, nil),
		temp:     prometheus.NewDesc(strings.Join([]string{"nest", "current", "temperature", unit}, "_"), "Inside temperature.", nestLabels, nil),
		target:   prometheus.NewDesc(strings.Join([]string{"nest", "target", "temperature", unit}, "_"), "Target temperature.", nestLabels, nil),
		humidity: prometheus.NewDesc(strings.Join([]string{"nest", "humidity", "percent"}, "_"), "Inside humidity.", nestLabels, nil),
		heating:  prometheus.NewDesc(strings.Join([]string{"nest", "heating"}, "_"), "Is thermostat heating.", nestLabels, nil),
		leaf:     prometheus.NewDesc(strings.Join([]string{"nest", "leaf", "percent"}, "_"), "Is thermostat set to energy-saving temperature.", nestLabels, nil),
	}
}

// Describe implements the prometheus.Describe interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.up
	ch <- c.metrics.temp
	ch <- c.metrics.target
	ch <- c.metrics.humidity
	ch <- c.metrics.heating
	ch <- c.metrics.leaf
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	thermostats, err := c.getNestReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.metrics.up, prometheus.GaugeValue, 0)
		c.logger.Log("level", "error", "message", "Failed collecting Nest data", "stack", errors.WithStack(err))
		return
	}

	c.logger.Log("level", "debug", "message", "Successfully collected Nest data")

	ch <- prometheus.MustNewConstMetric(c.metrics.up, prometheus.GaugeValue, 1)

	for _, therm := range thermostats {
		labels := []string{therm.ID, strings.Replace(therm.Name, " ", "-", -1)}

		if c.unit == celsius {
			ch <- prometheus.MustNewConstMetric(c.metrics.temp, prometheus.GaugeValue, therm.TemperatureC, labels...)
			ch <- prometheus.MustNewConstMetric(c.metrics.target, prometheus.GaugeValue, therm.TargetC, labels...)
		} else {
			ch <- prometheus.MustNewConstMetric(c.metrics.temp, prometheus.GaugeValue, therm.TemperatureF, labels...)
			ch <- prometheus.MustNewConstMetric(c.metrics.target, prometheus.GaugeValue, therm.TargetF, labels...)
		}

		ch <- prometheus.MustNewConstMetric(c.metrics.humidity, prometheus.GaugeValue, therm.Humidity, labels...)
		ch <- prometheus.MustNewConstMetric(c.metrics.heating, prometheus.GaugeValue, b2f(therm.HVACState == "heating"), labels...)
		ch <- prometheus.MustNewConstMetric(c.metrics.leaf, prometheus.GaugeValue, b2f(therm.Leaf), labels...)
	}
}

func (c *Collector) getNestReadings() (thermostats []*Thermostat, err error) {
	res, err := c.client.Do(c.req)
	if err != nil {
		return nil, errors.Wrap(errFailedRequest, err.Error())
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(errFailedReadingBody, err.Error())
	}

	if res.StatusCode != 200 {
		return nil, errors.Wrap(errNon200Response, fmt.Sprintf("code: %d", res.StatusCode))
	}

	var data map[string]Thermostat
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errors.Wrap(errFailedUnmarshalling, err.Error())
	}

	for _, thermostat := range data {
		thermostats = append(thermostats, &thermostat)
	}

	return thermostats, nil
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
