package jsonutil_test

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"iter"

	"github.com/pomerium/datasource/internal/jsonutil"
)

func collect[T any](it iter.Seq2[T, error]) ([]T, error) {
	var result []T
	for v, err := range it {
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

type testCase struct {
	name     string
	input    any
	expected any
	fn       func(r io.Reader) (any, error)
}

func mkTest[T any](
	name string,
	input any,
	expected any,
	keys []string,
) testCase {
	return testCase{
		name:     name + " " + strings.Join(keys, "."),
		input:    input,
		expected: expected,
		fn: func(r io.Reader) (any, error) {
			it := jsonutil.StreamArrayReader[T](r, keys)
			result, err := collect(it)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
	}
}

func TestReader(t *testing.T) {
	t.Parallel()

	type S struct {
		A int
		B string
	}

	ts := []S{
		{1, "a"},
		{2, "b"},
		{3, "c"},
	}

	tests := []testCase{
		mkTest[int]("int", []int{1, 2, 3}, []int{1, 2, 3}, nil),
		mkTest[string]("string", []string{"a", "b", "c"}, []string{"a", "b", "c"}, nil),
		mkTest[S]("struct", ts, ts, nil),
		mkTest[int]("int", map[string]any{"nested": []int{1, 2, 3}}, []int{1, 2, 3}, []string{"nested"}),
		mkTest[int]("int", map[string]any{"nested": map[string]any{"two-level": []int{1, 2, 3}}}, []int{1, 2, 3}, []string{"nested", "two-level"}),
		mkTest[int]("int", map[string]any{
			"j1": []int{4, 5, 6},
			"j2": "v",
			"j3": map[string]any{
				"j4": []int{4, 5, 6},
			},
			"key1": map[string]any{
				"j5":              []int{4, 5, 6},
				"j6":              "v2",
				"with-other-keys": []int{1, 2, 3},
			},
		}, []int{1, 2, 3}, []string{"key1", "with-other-keys"}),
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			encoded, err := json.Marshal(tc.input)
			require.NoError(t, err)

			result, err := tc.fn(strings.NewReader(string(encoded)))
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}
