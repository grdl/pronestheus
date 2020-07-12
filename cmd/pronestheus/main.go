package main

import (
	"pronestheus/pkg"

	"gopkg.in/alecthomas/kingpin.v2"
)

var cfg = &pkg.Config{
	ListenAddr:      kingpin.Flag("listen-addr", "nanan").Default(":9999").String(),
	Timeout:         kingpin.Flag("scrape-timeout", "The time to wait for remote APIs to response, in miliseconds").Default("5000").Int(),
	NestURL:         kingpin.Flag("nest-api-url", "The Nest API URL").Default("https://developer-api.nest.com/devices/thermostats").String(),
	NestToken:       kingpin.Flag("nest-api-token", "The authorization token for Nest API").Required().String(),
	WeatherURL:      kingpin.Flag("weather-api-url", "The OpenWeatherMap URL").Default("http://api.openweathermap.org/data/2.5/weather").String(),
	WeatherToken:    kingpin.Flag("weather-api-token", "The authorization token for OpenWeatherMap API").Required().String(),
	WeatherLocation: kingpin.Flag("weather-api-location-id", "The location ID for OpenWeatherMap API. Defaults to Amsterdam").Default("2759794").String(),
}

func main() {
	kingpin.Parse()

	pkg.Run(cfg)
}
