package internal

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
