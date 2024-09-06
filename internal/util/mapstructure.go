package util

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

// DateTime is used by some API services to represent date and/or time
// in a non-standard manners, that may also be
type DateTime struct {
	tm     time.Time
	layout string
}

func (d *DateTime) Time() time.Time {
	return d.tm
}

var (
	_              = json.Marshaler(new(DateTime))
	DateType       = reflect.TypeOf(DateTime{})
	JSONNumberType = reflect.TypeOf(json.Number(""))
)

// MarshalJSON implements json.Marshaler
func (d DateTime) MarshalJSON() ([]byte, error) {
	if d.tm.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(d.tm.Format(d.layout))
}

// String returns string representation in the target layout
func (d DateTime) String() string {
	if d.tm.IsZero() {
		return ""
	}

	return d.tm.Format(d.layout)
}

// NewDateTime creates new date time object
func NewDateTime(tm time.Time, layout string) DateTime {
	return DateTime{tm, layout}
}

// DateTimeDecodeHook parses date time that's supplied in a non-standard layout
// if layout does not contain a time zone, a location need be provided
func DateTimeDecodeHook(layout string, location *time.Location) mapstructure.DecodeHookFuncType {
	return func(_, dstT reflect.Type, data interface{}) (interface{}, error) {
		if dstT != DateType {
			return data, nil
		}

		if data == nil {
			return nil, nil
		}

		txt, ok := data.(string)
		if !ok {
			return data, nil
		}

		var tm time.Time
		var err error
		if location == nil {
			tm, err = time.Parse(layout, txt)
		} else {
			tm, err = time.ParseInLocation(layout, txt, location)
		}
		if err != nil {
			return nil, err
		}

		return &DateTime{tm, layout}, nil
	}
}

// JSONNumberDecodeHook helps mapstructure to deal with
func JSONNumberDecodeHook(_, dstT reflect.Type, data interface{}) (interface{}, error) {
	if dstT != JSONNumberType {
		return data, nil
	}

	switch val := data.(type) {
	case string:
		return val, nil
	case float64:
		x := math.Trunc(val)
		if x != val {
			return nil, fmt.Errorf("JSON number must be represented as a whole number, got %f", val)
		}
		return fmt.Sprint(int(x)), nil
	case int64:
		return fmt.Sprint(val), nil
	default:
		return nil, fmt.Errorf("unsupported format %v", val)
	}
}
