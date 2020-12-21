package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

// WeatherServerMetric returns a mock OpenWeatherMap server which returns a valid response with temperature in Celsius.
func WeatherServerMetric() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, readFile(filepath.Join("weather_metric.json")))
	}))
}

// WeatherServerImperial returns a mock OpenWeatherMap server which returns a valid response with temperature in Fahrenheit.
func WeatherServerImperial() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, readFile(filepath.Join("weather_imperial.json")))
	}))
}

// WeatherServerMissingID returns a mock OpenWeatherMap server which returns an error due to missing location ID.
func WeatherServerMissingID() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, readFile(filepath.Join("weather_empty_id.json")))
	}))
}

// WeatherServerInvalidToken returns a mock OpenWeatherMap server which returns an error due to invalid authentication token.
func WeatherServerInvalidToken() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, readFile(filepath.Join("weather_invalid_token.json")))
	}))
}

// WeatherServerInvalidResponse returns a mock OpenWeatherMap server which returns an invalid JSON response.
func WeatherServerInvalidResponse() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, readFile(filepath.Join("weather_invalid.json")))
	}))
}

// NestServer returns a mock Nest server which returns a valid response.
func NestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, readFile(filepath.Join("nest_valid.json")))
	}))
}

// NestServerInvalidToken returns a mock Nest server which returns an error due to invalid authentication token.
func NestServerInvalidToken() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, readFile(filepath.Join("nest_invalid_token.json")))
	}))
}

// NestServerInvalidResponse returns a mock Nest server which returns an invalid JSON response.
func NestServerInvalidResponse() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, readFile(filepath.Join("nest_invalid.json")))
	}))
}

// readFile returns contents of a file from the testdata folder.
//
// `go test` always executes tests with working directory set to the source of the package being tested.
// Because of that, we need to find the path to the testdata dir to be able to use it in tests inside different packages.
//
// Ref:
// - https://dave.cheney.net/2016/05/10/test-fixtures-in-go
// - https://stackoverflow.com/a/38644571/1085632
//
func readFile(filename string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	bytes, err := ioutil.ReadFile(filepath.Join(basepath, "testdata", filename))
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// ValidToken returns a dummy oauth token which is always valid.
// Using this token won't trigger a call to refresh the access token.
func ValidToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "dummy token",
		TokenType:    "Bearer",
		RefreshToken: "dummy refresh",
		Expiry:       time.Time{},
	}
}
