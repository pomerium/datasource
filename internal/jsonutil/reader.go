package jsonutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
)

func StreamArrayReadAndClose[T any](r io.ReadCloser, keys []string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		StreamArrayReader[T](r, keys)(yield)
		_ = r.Close()
	}
}

// StreamArrayReader reads a JSON array from r and yields each element.
// keys is a list of keys hierarchy to traverse before reading the array.
func StreamArrayReader[T any](r io.Reader, keys []string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var v T
		decoder := json.NewDecoder(r)

		err := traverseKeys(decoder, keys)
		if err != nil {
			yield(v, err)
			return
		}

		err = readDelim(decoder, json.Delim('['))
		if err != nil {
			yield(v, err)
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

		err = readDelim(decoder, json.Delim(']'))
		if err != nil {
			yield(v, err)
			return
		}
	}
}

func readDelim(decoder *json.Decoder, delim json.Delim) error {
	tk, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("reading next token: %w", err)
	}
	if tk != delim {
		return fmt.Errorf("expected `%v`, got `%v`", delim, tk)
	}
	return nil
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
	err := readDelim(decoder, json.Delim('{'))
	if err != nil {
		return err
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
