package util_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pomerium/datasource/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate(t *testing.T) {
	est, err := time.LoadLocation("EST")
	assert.NoError(t, err, "load EST timezone")

	for _, tc := range []struct {
		json     string
		layout   string
		location string
		expect   time.Time
	}{
		{`{"date":"2022-04-21"}`, "2006-01-02", "EST", time.Date(2022, 4, 21, 0, 0, 0, 0, est)},
		{`{"date":"2022-04-21"}`, "2006-01-02", "UTC", time.Date(2022, 4, 21, 0, 0, 0, 0, time.UTC)},
		{`{"date":null}`, "2006-01-02", "EST", time.Time{}},
	} {
		t.Run(tc.json, func(t *testing.T) {
			var v struct {
				Date util.DateTime `json:"date" mapstructure:"date"`
			}
			location, err := time.LoadLocation(tc.location)
			require.NoError(t, err)

			dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				DecodeHook: util.DateTimeDecodeHook(tc.layout, location),
				Result:     &v,
			})
			require.NoError(t, err)

			m := make(map[string]interface{})
			require.NoError(t, json.Unmarshal([]byte(tc.json), &m))
			require.NoError(t, dec.Decode(m))

			assert.Equal(t, tc.expect, v.Date.Time())

			data, err := json.Marshal(v)
			require.NoError(t, err)
			assert.Equal(t, tc.json, string(data))
		})
	}
}
