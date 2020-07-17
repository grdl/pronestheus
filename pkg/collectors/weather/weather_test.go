package weather

import (
	"pronestheus/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidResponseCelsius(t *testing.T) {
	c, err := New(Config{
		APIURL: test.WeatherServerMetric().URL,
	})
	assert.NoError(t, err)

	weather, err := c.getWeatherReadings()
	assert.NoError(t, err)

	assert.Equal(t, weather.Temperature, float64(20.26))
	assert.Equal(t, weather.Humidity, float64(88))
	assert.Equal(t, weather.Pressure, float64(1021))
}
func TestValidResponseFahrenheit(t *testing.T) {
	c, err := New(Config{
		APIURL: test.WeatherServerImperial().URL,
	})
	assert.NoError(t, err)

	weather, err := c.getWeatherReadings()
	assert.NoError(t, err)

	assert.Equal(t, weather.Temperature, float64(68.36))
	assert.Equal(t, weather.Humidity, float64(88))
	assert.Equal(t, weather.Pressure, float64(1021))
}
