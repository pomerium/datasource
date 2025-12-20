package util

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type URL url.URL

var (
	_ = json.Marshaler(new(URL))
	_ = json.Unmarshaler(new(URL))
)

func (u *URL) URL() *url.URL {
	if u == nil {
		return nil
	}
	x := url.URL(*u)
	return &x
}

func (u *URL) String() string {
	if u == nil {
		return "null"
	}
	x := url.URL(*u)
	return x.String()
}

func (u *URL) MarshalJSON() ([]byte, error) {
	if u == nil {
		return nil, nil
	}
	x := url.URL(*u)
	return json.Marshal(x.String())
}

func (u *URL) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var txt string
	if err := json.Unmarshal(data, &txt); err != nil {
		return fmt.Errorf("decode json string: %w", err)
	}
	res, err := url.Parse(txt)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	*u = URL(*res)
	return nil
}
