package weather

import (
	"errors"
	"pronestheus/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerResponses(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
		want    *Weather
	}{
		{
			name:    "valid response celsius",
			url:     test.WeatherServerMetric().URL,
			wantErr: nil,
			want: &Weather{
				Humidity:    float64(88),
				Pressure:    float64(1021),
				Temperature: float64(20.26),
			},
		}, {
			name:    "valid response fahrenheit",
			url:     test.WeatherServerImperial().URL,
			wantErr: nil,
			want: &Weather{
				Humidity:    float64(88),
				Pressure:    float64(1021),
				Temperature: float64(68.36),
			},
		}, {
			name:    "missing location id",
			url:     test.WeatherServerMissingID().URL,
			wantErr: errNon200Response,
			want:    nil,
		}, {
			name:    "invalid auth token",
			url:     test.WeatherServerInvalidToken().URL,
			wantErr: errNon200Response,
			want:    nil,
		}, {
			name:    "invalid JSON response",
			url:     test.WeatherServerInvalidResponse().URL,
			wantErr: errFailedUnmarshalling,
			want:    nil,
		}, {
			name:    "invalid server",
			url:     "http://nonexisting.server",
			wantErr: errFailedRequest,
			want:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := New(Config{
				APIURL: test.url,
			})
			assert.NoError(t, err)

			weather, err := c.getWeatherReadings()

			if test.wantErr != nil {
				assert.Nil(t, weather)
				assert.True(t, errors.Is(err, test.wantErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, weather, test.want)
			}
		})
	}
}

func TestAPIURLParsing(t *testing.T) {
	tests := []struct {
		name    string
		rawurl  string
		wantErr error
	}{
		{
			name:    "invalid url",
			rawurl:  "https/////this.is.not.a.valid.url",
			wantErr: errFailedParsingURL,
		}, {
			name:    "empty url",
			rawurl:  "",
			wantErr: errFailedParsingURL,
		}, {
			name:    "valid url",
			rawurl:  "https://example.com/valid",
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := New(Config{
				APIURL: test.rawurl,
			})

			if test.wantErr != nil {
				assert.Nil(t, c)
				assert.True(t, errors.Is(err, test.wantErr))
			} else {
				assert.NotNil(t, c)
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIURLUnits(t *testing.T) {
	tests := []struct {
		name    string
		unit    string
		wantURL string
		wantErr error
	}{
		{
			name:    "valid celsius",
			unit:    "celsius",
			wantURL: "https://example.com?id=123&appid=abc&units=metric",
			wantErr: nil,
		}, {
			name:    "valid fahrenheit",
			unit:    "fahrenheit",
			wantURL: "https://example.com?id=123&appid=abc&units=imperial",
			wantErr: nil,
		}, {
			name:    "valid empty",
			unit:    "",
			wantURL: "https://example.com?id=123&appid=abc&units=metric",
			wantErr: nil,
		}, {
			name:    "invalid",
			unit:    "furlong",
			wantURL: "",
			wantErr: errInvalidTempUnit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := New(Config{
				APIURL:        "https://example.com",
				APILocationID: "123",
				APIToken:      "abc",
				Unit:          test.unit,
			})

			if test.wantErr != nil {
				assert.Nil(t, c)
				assert.True(t, errors.Is(err, test.wantErr))
			} else {
				assert.Equal(t, c.url, test.wantURL)
				assert.NoError(t, err)
			}
		})
	}
}
