package weather

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
	errNon200Response      = errors.New("openWeatherMap API responded with non-200 code")
	errFailedParsingURL    = errors.New("failed parsing OpenWeatherMap API URL")
	errInvalidTempUnit     = errors.New("invalid temperature unit; valid values: [celsius, fahrenheit]")
	errFailedUnmarshalling = errors.New("failed unmarshalling OpenWeatherMap API response body")
	errFailedRequest       = errors.New("failed OpenWeatherMap API request")
	errFailedReadingBody   = errors.New("failed reading OpenWeatherMap API response body")
)

// Weather stores weather data received from OpenWeatherMap API.
type Weather struct {
	Temperature float64 `json:"temp"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}

// Config provides the configuration necessary to create the Collector.
type Config struct {
	Logger        log.Logger
	Timeout       int
	Unit          string
	APIURL        string
	APIToken      string
	APILocationID string
}

// Collector implements the Collector interface, collecting weather data from OpenWeatherMap API.
type Collector struct {
	client  *http.Client
	url     string
	logger  log.Logger
	metrics *Metrics
}

// Metrics contains the metrics collected by the Collector.
type Metrics struct {
	up       *prometheus.Desc
	temp     *prometheus.Desc
	humidity *prometheus.Desc
	pressure *prometheus.Desc
}

// New creates a Collector using the given Config.
func New(cfg Config) (*Collector, error) {
	var units string
	switch cfg.Unit {
	case "", celsius:
		units = "metric"
	case fahrenheit:
		units = "imperial"
	default:
		return nil, errInvalidTempUnit
	}

	rawurl := fmt.Sprintf("%s?id=%s&appid=%s&units=%s", cfg.APIURL, cfg.APILocationID, cfg.APIToken, units)
	if _, err := url.ParseRequestURI(rawurl); err != nil {
		return nil, errors.Wrap(errFailedParsingURL, err.Error())
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Millisecond,
	}

	collector := &Collector{
		client:  client,
		url:     rawurl,
		logger:  cfg.Logger,
		metrics: buildMetrics(cfg.Unit),
	}

	return collector, nil
}

func buildMetrics(unit string) *Metrics {
	if unit == "" {
		unit = "celsius"
	}

	return &Metrics{
		up:       prometheus.NewDesc(strings.Join([]string{"nest", "weather", "up"}, "_"), "Was talking to OpenWeatherMap API successful.", nil, nil),
		temp:     prometheus.NewDesc(strings.Join([]string{"nest", "weather", "temperature", unit}, "_"), "Outside temperature.", nil, nil),
		humidity: prometheus.NewDesc(strings.Join([]string{"nest", "weather", "humidity", "percent"}, "_"), "Outside humidity.", nil, nil),
		pressure: prometheus.NewDesc(strings.Join([]string{"nest", "weather", "pressure", "hectopascal"}, "_"), "Outside pressure.", nil, nil),
	}
}

// Describe implements the prometheus.Describe interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.up
	ch <- c.metrics.temp
	ch <- c.metrics.humidity
	ch <- c.metrics.pressure
}

// Collect implements the prometheus.Describe interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	weather, err := c.getWeatherReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.metrics.up, prometheus.GaugeValue, 0)
		c.logger.Log("level", "error", "message", "Failed collecting OpenWeatherMap data", "stack", errors.WithStack(err))
		return
	}

	c.logger.Log("level", "debug", "message", "Successfully collected OpenWeatherMap data")

	ch <- prometheus.MustNewConstMetric(c.metrics.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.metrics.temp, prometheus.GaugeValue, weather.Temperature)
	ch <- prometheus.MustNewConstMetric(c.metrics.humidity, prometheus.GaugeValue, weather.Humidity)
	ch <- prometheus.MustNewConstMetric(c.metrics.pressure, prometheus.GaugeValue, weather.Pressure)
}

func (c *Collector) getWeatherReadings() (weather *Weather, err error) {
	res, err := c.client.Get(c.url)
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

	var data map[string]json.RawMessage

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errors.Wrap(errFailedUnmarshalling, err.Error())
	}

	err = json.Unmarshal(data["main"], &weather)
	if err != nil {
		return nil, errors.Wrap(errFailedUnmarshalling, err.Error())
	}

	return weather, nil
}
