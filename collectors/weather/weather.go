package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/micrologger"
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
	Logger micrologger.Logger

	ApiURL        string
	ApiToken      string
	ApiLocationID string
}

// Collector implements the Collector interface, collecting weather data from OpenWeatherMap APi
type Collector struct {
	logger micrologger.Logger

	apiURL        string
	apiToken      string
	apiLocationID string

	up       *prometheus.Desc
	temp     *prometheus.Desc
	humidity *prometheus.Desc
	pressure *prometheus.Desc

	// errors?
}

// New creates a Collector with given Config
func New(config Config) (*Collector, error) {
	//if config.ApiURL == nil {
	//	return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	//}
	//if config.ApiToken == nil {
	//	return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	//}
	//if config.ApiLocationID == nil {
	//	return nil, microerror.Maskf(invalidConfigError, "%T.TCPClient must not be empty", config)
	//}

	collector := &Collector{
		logger:        config.Logger,
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
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.temp
	ch <- c.humidity
	ch <- c.pressure
}

// Collect implements the Collect method of the Collector interface.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "Scraping OpenWeatherMap API")

	weather, err := c.getWeatherReadings()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		c.logger.Log("level", "error", "message", "could not get weather readings", "stack", microerror.Stack(err))
		return
	}

	c.logger.Log("level", "debug", "message", "Successfully scraped OpenWeatherMap API")

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.temp, prometheus.GaugeValue, weather.Temperature)
	ch <- prometheus.MustNewConstMetric(c.humidity, prometheus.GaugeValue, weather.Humidity)
	ch <- prometheus.MustNewConstMetric(c.pressure, prometheus.GaugeValue, weather.Pressure)
}

func (c Collector) getWeatherReadings() (weather Weather, err error) {
	url := fmt.Sprintf("%s?id=%s&appid=%s&units=metric", c.apiURL, c.apiLocationID, c.apiToken)

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