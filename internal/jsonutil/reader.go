package jsonutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
)

// StreamArrayReader reads a JSON array from r and yields each element.
func StreamArrayReader[T any](r io.Reader, keys []string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var v T
		decoder := json.NewDecoder(r)

		err := traverseKeys(decoder, keys)
		if err != nil {
			yield(v, err)
			return
		}

		tk, err := decoder.Token()
		if err != nil {
			yield(v, fmt.Errorf("reading next token: %w", err))
			return
		}
		if tk != json.Delim('[') {
			yield(v, fmt.Errorf("expected `[`, got `%v`", tk))
			return
		}

		for decoder.More() {
			err := decoder.Decode(&v)
			if errors.Is(err, io.EOF) {
				break
			}
			if !yield(v, err) {
				return
			}
			if err != nil {
				return
			}
		}

		tk, err = decoder.Token()
		if err != nil {
			yield(v, fmt.Errorf("reading next token: %w", err))
			return
		}
		if tk != json.Delim(']') {
			yield(v, fmt.Errorf("expected `[`, got `%v`", tk))
			return
		}
	}
}

func skipObject(decoder *json.Decoder) error {
	t, err := decoder.Token()
	if err != nil {
		return err
	}

	_, ok := t.(json.Delim)
	if !ok {
		return nil
	}

	for decoder.More() {
		err := skipObject(decoder)
		if err != nil {
			return err
		}
	}

	t, err = decoder.Token()
	if err != nil {
		return err
	}
	_, ok = t.(json.Delim)
	if !ok {
		return fmt.Errorf("expected delimeter, got %v", t)
	}

	return nil
}

func findObjectKey(decoder *json.Decoder, key string) error {
	t, err := decoder.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("expected a {, got %v", t)
	}

	for {
		t, err := decoder.Token()
		if err != nil {
			return err
		}
		keyName, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected a string key, got %v", t)
		}
		if keyName == key {
			return nil
		}
		err = skipObject(decoder)
		if err != nil {
			return fmt.Errorf("error skipping object for key %s: %w", keyName, err)
		}
	}
}

func traverseKeys(decoder *json.Decoder, keys []string) error {
	for _, key := range keys {
		err := findObjectKey(decoder, key)
		if err != nil {
			return err
		}
	}
	return nil
}
