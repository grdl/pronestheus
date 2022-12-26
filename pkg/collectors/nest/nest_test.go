package nest

import (
	mock "pronestheus/test"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/pkg/errors"
)

func TestServerResponses(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
		want    *Thermostat
	}{
		{
			name:    "valid response",
			url:     mock.NestServer().URL,
			wantErr: nil,
			want: &Thermostat{
				ID:           "enterprises/PROJECT_ID/devices/DEVICE_ID",
				Label:        "Custom Name",
				Online:       true,
				AmbientTemp:  float64(20.23999),
				SetpointTemp: float64(19.17838),
				Humidity:     float64(57),
				Status:       "OFF",
			},
		}, {
			name:    "invalid auth token",
			url:     mock.NestServerInvalidToken().URL,
			wantErr: errNon200Response,
			want:    nil,
		}, {
			name:    "invalid JSON response",
			url:     mock.NestServerInvalidResponse().URL,
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
				APIURL:     test.url,
				OAuthToken: mock.ValidToken(),
			})
			assert.NoError(t, err)

			thermostats, err := c.getNestReadings()

			if test.wantErr != nil {
				assert.Nil(t, thermostats)
				assert.True(t, errors.Is(err, test.wantErr))
			} else {
				assert.NoError(t, err)
				assert.Len(t, thermostats, 1)
				assert.Equal(t, thermostats[0], test.want)
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
