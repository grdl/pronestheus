package main

import (
	"errors"

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

var (
	nestUp = prometheus.NewDesc(
		"pns_nest_up",
		"Was talking to Nest API successful.",
		nil, nil,
	)
)

type NestCollector struct {
}

// Implements prometheus.Collector
func (c NestCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nestUp
}

// Implements prometheus.Collector
func (c NestCollector) Collect(ch chan<- prometheus.Metric) {
	thermostats, err := getNestData()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(nestUp, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(nestUp, prometheus.GaugeValue, 1)

	log.Info(thermostats)
}

func getNestData() (result string, err error) {

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
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {
	kingpin.Parse()

	c := NestCollector{}
	prometheus.MustRegister(c)

	log.With("addr", *listenAddress).Info("Started listening")

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
