package weather

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// Weather stores weather data received from OpenWeatherMap API
type Weather struct {
	Temperature float64 `json:"temp"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}

// Config provides the necessary configuration for creating a Collector
type Config struct {
	Logger  log.Logger
	Timeout int

	ApiURL        string
	ApiToken      string
	ApiLocationID string
}

// Collector implements the Collector interface, collecting weather data from OpenWeatherMap APi
type Collector struct {
	logger  log.Logger
	timeout time.Duration

	apiURL        string
	apiToken      string
	apiLocationID string

	up       *prometheus.Desc
	temp     *prometheus.Desc
	humidity *prometheus.Desc
	pressure *prometheus.Desc

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
		return nil, errors.New("OpenWeatherMap Api URL config must not be empty")
	}
	if config.ApiToken == "" {
		return nil, errors.New("OpenWeatherMap Api Token config must not be empty")
	}
	if config.ApiLocationID == "" {
		return nil, errors.New("OpenWeatherMap Api Location ID config must not be empty")
	}

	collector := &Collector{
		logger:        config.Logger,
		timeout:       time.Duration(config.Timeout) * time.Millisecond,
		apiURL:        config.ApiURL,
		apiToken:      config.ApiToken,
		apiLocationID: config.ApiLocationID,
		up:            prometheus.NewDesc("nest_weather_up", "Was talking to OpenWeatherMap API successful.", nil, nil),
		temp:          prometheus.NewDesc("nest_weather_temp", "Current outside temperature.", nil, nil),
		humidity:      prometheus.NewDesc("nest_weather_humidity", "Current outside humidity", nil, nil),
		pressure:      prometheus.NewDesc("nest_weather_pressure", "Current outside pressure", nil, nil),
	}

	return collector, nil
}

// Describe implements the Describe method of the Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.temp
	ch <- c.humidity
	ch <- c.pressure
}

// Collect implements the Collect method of the Collector interface.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "Scraping OpenWeatherMap API")

	weather, err := c.getWeatherReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		c.logger.Log("level", "error", "message", "could not get weather readings", "stack", errors.WithStack(err))
		return
	}

	c.logger.Log("level", "debug", "message", "Successfully scraped OpenWeatherMap API")

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.temp, prometheus.GaugeValue, weather.Temperature)
	ch <- prometheus.MustNewConstMetric(c.humidity, prometheus.GaugeValue, weather.Humidity)
	ch <- prometheus.MustNewConstMetric(c.pressure, prometheus.GaugeValue, weather.Pressure)
}

func (c *Collector) getWeatherReadings() (weather *Weather, err error) {
	url := fmt.Sprintf("%s?id=%s&appid=%s&units=metric", c.apiURL, c.apiLocationID, c.apiToken)
	req, _ := http.NewRequest("GET", url, nil)

	client := http.Client{
		Timeout: c.timeout,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Calling OpenWeatherMap API failed")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Reading OpenWeatherMap API response failed")
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("OpenWeatherMap responded with %d code: %s", res.StatusCode, body))
	}

	var data map[string]json.RawMessage

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshalling OpenWeatherMap API response failed")
	}

	err = json.Unmarshal(data["main"], &weather)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshalling OpenWeatherMap API response failed")
	}

	return weather, nil
}
