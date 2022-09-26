package util

import (
	"fmt"
	"strings"
)

type FieldRemap struct {
	From, To string
}

// NewRemapFromPairs takes an array of srcKey=dstKey pairs and converts them into FieldRemap array
func NewRemapFromPairs(src []string) ([]FieldRemap, error) {
	dst := make([]FieldRemap, 0, len(src))
	keys := make(map[string]struct{})

	for _, r := range src {
		pair := strings.Split(r, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("%s: expect key=newKey format", r)
		}
		fm := FieldRemap{From: pair[0], To: pair[1]}
		if fm.From == "" || fm.To == "" {
			return nil, fmt.Errorf("%s: expect key=newKey format", r)
		}
		if _, there := keys[fm.To]; there {
			return nil, fmt.Errorf("%s: key %s was already used", r, fm.To)
		}
		dst = append(dst, fm)
		keys[fm.To] = struct{}{}
	}
	return dst, nil
}

// Remap is used to rename map keys in-place
func Remap(src []map[string]interface{}, fieldMap []FieldRemap) error {
	for _, m := range src {
		if err := remap(m, fieldMap); err != nil {
			return fmt.Errorf("%+v: %w", m, err)
		}
	}
	return nil
}

// Filter removes fields that are not part of the list
func Filter(src []map[string]interface{}, fields []string) {
	keep := make(map[string]struct{})
	for _, field := range fields {
		keep[field] = struct{}{}
	}

	for _, m := range src {
		for k := range m {
			if _, there := keep[k]; !there {
				delete(m, k)
			}
		}
	}
}

func remap(src map[string]interface{}, fieldMap []FieldRemap) error {
	for _, fm := range fieldMap {
		val, there := src[fm.From]
		if !there {
			continue
		}

		if _, there = src[fm.To]; there {
			return fmt.Errorf("cannot rename key %s to %s: already exists", fm.From, fm.To)
		}
		delete(src, fm.From)
		src[fm.To] = val
	}
	return nil
}
