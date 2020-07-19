# ProNestheus

![build](https://github.com/grdl/pronestheus/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/grdl/pronestheus)](https://goreportcard.com/report/github.com/grdl/pronestheus)

A Prometheus exporter for the [Nest Learning Thermostat](https://nest.com/).

Exposes metrics about your thermostats and weather in your current location.

![dashboard](docs/dashboard.png)

## Installation

### Binary download

Grab the Linux, macOS or Windows executable from the [latest release](https://github.com/grdl/pronestheus/releases/latest).

### Docker image

```
docker run -p 9777:9777 -e "PRONESTHEUS_NEST_TOKEN=xxx" grdl/pronestheus
```

### Helm chart

Helm chart is available in `deployments/helm`.

### "One-click" installation with Docker Compose

Update necessary variables in `deployments/docker-compose/.env` file. Then run:
```
cd deployments/docker-compose
docker-compose up
```

This will start docker containers with Prometheus, Grafana and ProNestheus exporter automatically configured. Visit http://localhost:3000 to open Grafana dashboard with your thermostat metrics.


### Usage and configuration

```
usage: pronestheus --nest-auth=NEST-AUTH [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --listen-addr              Address on which to expose metrics and web interface. (default ":9777")
      --metrics-path             Path under which to expose metrics. (default "/metrics")
      --scrape-timeout           Time to wait for remote APIs to response, in milliseconds. (default 5000)
      --temp-unit                Temperature metric unit [celsius, fahrenheit]. (default "celsius")
      --nest-url                 Nest API URL. (default "https://developer-api.nest.com/devices/thermostats")
      --nest-auth [mandatory]    Authorization token for Nest API.
      --owm-url                  OpenWeatherMap API URL (default "http://api.openweathermap.org/data/2.5/weather")
      --owm-auth                 Authorization token for OpenWeatherMap API.
      --owm-location             Location ID for OpenWeatherMap API. Defaults to Amsterdam. (default "2759794")      
  -v, --version                  Show application version.
```

If `--owm-location` is not provided, the weather metrics are not exported.

All configuration flags can be passed as environment variables with `PRONESTHEUS_` prefix. Eg, `PRONESTHEUS_NEST_AUTH`.


### Authentication

Nest API token is required to call Nest API.

OpenWeatherMap API key is required to call the weather API. [Look here](https://openweathermap.org/appid) for instructions on how to get it.


## Exported metrics

```
# HELP nest_current_temperature_celsius Inside temperature.
# TYPE nest_current_temperature_celsius gauge
nest_current_temperature_celsius{id="abcd1234",name="Living-Room"} 23.5
# HELP nest_heating Is thermostat heating.
# TYPE nest_heating gauge
nest_heating{id="abcd1234",name="Living-Room"} 0
# HELP nest_humidity_percent Inside humidity.
# TYPE nest_humidity_percent gauge
nest_humidity_percent{id="abcd1234",name="Living-Room"} 55
# HELP nest_leaf Is thermostat set to energy-saving temperature.
# TYPE nest_leaf gauge
nest_leaf{id="abcd1234",name="Living-Room"} 1
# HELP nest_target_temperature_celsius Target temperature.
# TYPE nest_target_temperature_celsius gauge
nest_target_temperature_celsius{id="abcd1234",name="Living-Room"} 18
# HELP nest_up Was talking to Nest API successful.
# TYPE nest_up gauge
nest_up 1
# HELP nest_weather_humidity_percent Outside humidity.
# TYPE nest_weather_humidity_percent gauge
nest_weather_humidity_percent 82
# HELP nest_weather_pressure_hectopascal Outside pressure.
# TYPE nest_weather_pressure_hectopascal gauge
nest_weather_pressure_hectopascal 1016
# HELP nest_weather_temperature_celsius Outside temperature.
# TYPE nest_weather_temperature_celsius gauge
nest_weather_temperature_celsius 17.57
# HELP nest_weather_up Was talking to OpenWeatherMap API successful.
# TYPE nest_weather_up gauge
nest_weather_up 1
```
