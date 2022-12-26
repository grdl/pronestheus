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
	assert.Contains(t, w.Body.String(), `nest_online{id="enterprises/PROJECT_ID/devices/DEVICE_ID",label="Custom-Name"} 1`)
	assert.Contains(t, w.Body.String(), `nest_setpoint_temperature_celsius{id="enterprises/PROJECT_ID/devices/DEVICE_ID",label="Custom-Name"} 19.17838`)
	assert.Contains(t, w.Body.String(), `nest_ambient_temperature_celsius{id="enterprises/PROJECT_ID/devices/DEVICE_ID",label="Custom-Name"} 20.23999`)
	assert.Contains(t, w.Body.String(), `nest_humidity_percent{id="enterprises/PROJECT_ID/devices/DEVICE_ID",label="Custom-Name"} 57`)
	assert.Contains(t, w.Body.String(), `nest_heating{id="enterprises/PROJECT_ID/devices/DEVICE_ID",label="Custom-Name"} 0`)
	assert.Contains(t, w.Body.String(), "nest_weather_up 1")
	assert.Contains(t, w.Body.String(), "nest_weather_temperature_celsius 20.26")
	assert.Contains(t, w.Body.String(), "nest_weather_humidity_percent 88")
	assert.Contains(t, w.Body.String(), "nest_weather_pressure_hectopascal 1021")

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
	// Using dummy value to avoid nil-reference errors when creating test collectors.
	dummy := "dummy"

	return &ExporterConfig{
		ListenAddr:            &listenAddr,
		MetricsPath:           &metricsPath,
		Timeout:               &timeout,
		NestURL:               &dummy,
		NestOAuthClientID:     &dummy,
		NestOAuthClientSecret: &dummy,
		NestProjectID:         &dummy,
		NestRefreshToken:      &dummy,
		NestOAuthToken:        test.ValidToken(),
		WeatherLocation:       &dummy,
		WeatherURL:            &dummy,
		WeatherToken:          &dummy,
	}
}

// resetRegistry resets the default registry of Prometheus after each test.
// Without it, subsequent tests will fail because metrics were already registered in previous tests.
func resetRegistry() {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg
}
