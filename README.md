# ProNestheus

![build](https://github.com/grdl/pronestheus/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/grdl/pronestheus)](https://goreportcard.com/report/github.com/grdl/pronestheus)


A Prometheus exporter for Nest Learning Thermostat.

Exports metrics about your thermostats via Nest Developer API and weather metrics from your current location via OpenWeatherMap API. 

## Exported thermostat metrics

- `nest_up` - Was talking to Nest API successful
- `nest_current_temp` - Current ambient temperature
- `nest_target_temp` - Current target temperature
- `nest_humidity` - Current inside humidity
- `nest_heating` - Is thermostat heating
- `nest_leaf` - Is thermostat set to energy-saving temperature
- `nest_sunlight` - Is thermostat in direct sunlight

## Exporter weather metrics

- `nest_weather_up` - Was talking to OpenWeatherMap API successful
- `nest_weather_temp` - Current outside temperature
- `nest_weather_humidity` - Current outside humidity
- `nest_weather_pressure` - Current outside pressure


# Usage
```
usage: pronestheus --nest-api-token=NEST-API-TOKEN [<flags>]

Flags:
  --help                  Show context-sensitive help (also try --help-long and --help-man).
  --listen-addr=":2112"   The address to listen on
  --nest-api-url="https://developer-api.nest.com/devices/thermostats"
                          The Nest API URL
  --nest-api-token=NEST-API-TOKEN
                          The authorization token for Nest API
  --weather-api-url="http://api.openweathermap.org/data/2.5/weather"
                          The OpenWeatherMap URL
  --weather-api-token=""  The authorization token for OpenWeatherMap API
  --weather-api-location-id="2759794"
                          The location ID for OpenWeatherMap API. Defaults to Amsterdam
```