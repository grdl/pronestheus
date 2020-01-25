package main

import (
	"fmt"

	"gitlab.com/grdl/pronestheus/collectors/nest"
	"gitlab.com/grdl/pronestheus/collectors/weather"

	"github.com/giantswarm/exporterkit"
	"github.com/giantswarm/micrologger"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("listen-addr", "The address to listen on").Default("").String()
	scrapeTimeout = kingpin.Flag("scrape-timeout", "The time to wait for remote APIs to response, in miliseconds").Default("5000").Int()

	nestApiURL   = kingpin.Flag("nest-api-url", "The Nest API URL").Default("https://developer-api.nest.com/devices/thermostats").String()
	nestApiToken = kingpin.Flag("nest-api-token", "The authorization token for Nest API").Required().String()

	weatherApiURL        = kingpin.Flag("weather-api-url", "The OpenWeatherMap URL").Default("http://api.openweathermap.org/data/2.5/weather").String()
	weatherApiToken      = kingpin.Flag("weather-api-token", "The authorization token for OpenWeatherMap API").Required().String()
	weatherApiLocationId = kingpin.Flag("weather-api-location-id", "The location ID for OpenWeatherMap API. Defaults to Amsterdam").Default("2759794").String()
)

func main() {
	kingpin.Parse()

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}

	nestConfig := nest.Config{
		Logger:   logger,
		Timeout:  *scrapeTimeout,
		ApiURL:   *nestApiURL,
		ApiToken: *nestApiToken,
	}

	nestCollector, err := nest.New(nestConfig)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}

	weatherConfig := weather.Config{
		Logger:        logger,
		Timeout:       *scrapeTimeout,
		ApiURL:        *weatherApiURL,
		ApiToken:      *weatherApiToken,
		ApiLocationID: *weatherApiLocationId,
	}

	weatherCollector, err := weather.New(weatherConfig)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}

	exporterConfig := exporterkit.Config{
		Address: *listenAddress,
		Logger:  logger,
		Collectors: []prometheus.Collector{
			nestCollector,
			weatherCollector,
		},
	}

	exporter, err := exporterkit.New(exporterConfig)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}

	exporter.Run()
}
