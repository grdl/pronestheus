package nest

import (
	"pronestheus/test"
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
			url:     test.NestServer().URL,
			wantErr: nil,
			want: &Thermostat{
				ID:           "abcd1234567890",
				Name:         "Living Room",
				TemperatureC: float64(23.0),
				TemperatureF: float64(74),
				TargetC:      float64(20.0),
				TargetF:      float64(68),
				Humidity:     float64(60),
				HVACState:    "off",
				Leaf:         false,
			},
		}, {
			name:    "invalid auth token",
			url:     test.NestServerInvalidToken().URL,
			wantErr: errNon200Response,
			want:    nil,
		}, {
			name:    "invalid JSON response",
			url:     test.NestServerInvalidResponse().URL,
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