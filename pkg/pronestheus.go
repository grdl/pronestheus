package pkg

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pronestheus/pkg/collectors/nest"
	"pronestheus/pkg/collectors/weather"

	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	ListenAddr      *string
	Timeout         *int
	NestToken       *string
	NestURL         *string
	WeatherLocation *string
	WeatherURL      *string
	WeatherToken    *string
}

func Run(cfg *Config) {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	nestConfig := nest.Config{
		Logger:   logger,
		Timeout:  *cfg.Timeout,
		ApiURL:   *cfg.NestURL,
		ApiToken: *cfg.NestToken,
	}

	nestCollector, err := nest.New(nestConfig)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}

	weatherConfig := weather.Config{
		Logger:        logger,
		Timeout:       *cfg.Timeout,
		ApiURL:        *cfg.WeatherURL,
		ApiToken:      *cfg.WeatherToken,
		ApiLocationID: *cfg.WeatherLocation,
	}

	weatherCollector, err := weather.New(weatherConfig)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}

	prometheus.MustRegister(nestCollector)
	prometheus.MustRegister(weatherCollector)

	logger.Log("level", "debug", "msg", "Started Pronestheus - Nest Thermostat Prometheus Exporter")

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(*cfg.ListenAddr, nil)
	if err != nil {
		logger.Log("level", "error", "msg", err.Error())
	}

}
