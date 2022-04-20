package util

import "fmt"

// Remap is used to rename map keys in-place
func Remap(src []map[string]interface{}, fieldMap map[string]string) error {
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

func remap(src map[string]interface{}, fieldMap map[string]string) error {
	for key, val := range src {
		targetKey, there := fieldMap[key]
		if !there {
			continue
		}
		if _, there = src[targetKey]; there {
			return fmt.Errorf("cannot rename key %s to %s: already exists", key, targetKey)
		}
		delete(src, key)
		src[targetKey] = val
	}
	return nil
}
