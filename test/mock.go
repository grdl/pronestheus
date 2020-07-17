package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
)

// WeatherServerMetric returns a mock OpenWeatherMap server which returns valid response with temperature in Celsius.
func WeatherServerMetric() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, readFile(filepath.Join("weather_metric.json")))
	}))
}

// WeatherServerImperial returns a mock OpenWeatherMap server which returns valid response with temperature. in Fahrenheit.
func WeatherServerImperial() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, readFile(filepath.Join("weather_imperial.json")))
	}))
}

// readFile returns contents of a file from the testdata folder.
//
// Becasue `go test` always executes tests with working directory set to the source of the package under test,
// we need to find the path to the testdata dir to be able to use it in tests inside different packages.
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
