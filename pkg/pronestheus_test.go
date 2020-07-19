package pkg

import (
	"net/http"
	"net/http/httptest"
	"pronestheus/test"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func TestAllMetrics(t *testing.T) {
	t.Cleanup(resetRegistry)

	nestServ := test.NestServer()
	weatherServ := test.WeatherServerMetric()

	cfg := testConfig()
	cfg.NestURL = &nestServ.URL
	cfg.WeatherURL = &weatherServ.URL

	_, err := NewExporter(cfg)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	promhttp.Handler().ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), "nest_up 1")
	assert.Contains(t, w.Body.String(), `nest_target_temperature_celsius{id="abcd1234567890",name="Living-Room"} 20`)
	assert.Contains(t, w.Body.String(), `nest_current_temperature_celsius{id="abcd1234567890",name="Living-Room"} 23`)
	assert.Contains(t, w.Body.String(), `nest_humidity_percent{id="abcd1234567890",name="Living-Room"} 60`)
	assert.Contains(t, w.Body.String(), `nest_leaf{id="abcd1234567890",name="Living-Room"} 0`)
	assert.Contains(t, w.Body.String(), `nest_heating{id="abcd1234567890",name="Living-Room"} 0`)
	assert.Contains(t, w.Body.String(), "nest_weather_up 1")
	assert.Contains(t, w.Body.String(), "nest_weather_temperature_celsius 20.26")
	assert.Contains(t, w.Body.String(), "nest_weather_humidity_percent 88")
	assert.Contains(t, w.Body.String(), "nest_weather_pressure_hectopascal 1021")

}

func TestFahrenheitMetrics(t *testing.T) {
	t.Cleanup(resetRegistry)

	unit := "fahrenheit"
	nestServ := test.NestServer()
	weatherServ := test.WeatherServerImperial()

	cfg := testConfig()
	cfg.NestURL = &nestServ.URL
	cfg.WeatherURL = &weatherServ.URL
	cfg.TemperatureUnit = &unit

	_, err := NewExporter(cfg)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	promhttp.Handler().ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), "nest_up 1")
	assert.Contains(t, w.Body.String(), `nest_target_temperature_fahrenheit{id="abcd1234567890",name="Living-Room"} 68`)
	assert.Contains(t, w.Body.String(), `nest_current_temperature_fahrenheit{id="abcd1234567890",name="Living-Room"} 74`)
	assert.Contains(t, w.Body.String(), "nest_weather_up 1")
	assert.Contains(t, w.Body.String(), "nest_weather_temperature_fahrenheit 68.36")
}
func TestNoWeatherMetrics(t *testing.T) {
	t.Cleanup(resetRegistry)

	weatherToken := ""
	nestServ := test.NestServer()

	cfg := testConfig()
	cfg.NestURL = &nestServ.URL
	cfg.WeatherToken = &weatherToken

	_, err := NewExporter(cfg)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	promhttp.Handler().ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), "nest_up 1")
	assert.NotContains(t, w.Body.String(), "nest_weather_up 1")
}

func TestFailedScraping(t *testing.T) {
	t.Cleanup(resetRegistry)

	nestServ := test.NestServerInvalidResponse()
	weatherServ := test.WeatherServerInvalidResponse()

	cfg := testConfig()
	cfg.NestURL = &nestServ.URL
	cfg.WeatherURL = &weatherServ.URL

	_, err := NewExporter(cfg)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	promhttp.Handler().ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.NotContains(t, w.Body.String(), "nest_up 1")
	assert.NotContains(t, w.Body.String(), "nest_weather_up 1")
}

func testConfig() *ExporterConfig {
	listenAddr := ":9999"
	metricsPath := "/metrics"
	timeout := 5000
	unit := "celsius"
	nestToken := "abc"
	weatherLocation := "0"
	weatherToken := "abc"

	return &ExporterConfig{
		NestURL:         nil,
		NestToken:       &nestToken,
		WeatherURL:      nil,
		WeatherLocation: &weatherLocation,
		WeatherToken:    &weatherToken,
		ListenAddr:      &listenAddr,
		MetricsPath:     &metricsPath,
		TemperatureUnit: &unit,
		Timeout:         &timeout,
	}
}

// resetRegistry resets the default registry of Prometheus after each test.
// Without it, subsequent tests will fail because metrics were already registered in previous tests.
func resetRegistry() {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg
}
