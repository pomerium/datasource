package util

import (
	"reflect"
	"strings"
)

// GetStructTags returns list of names for a given struct tag
func GetStructTagNames(val any, tag string) []string {
	fields := reflect.VisibleFields(reflect.Indirect(reflect.ValueOf(val)).Type())
	names := make([]string, 0, len(fields))

	for _, field := range fields {
		name, there := field.Tag.Lookup(tag)
		if !there {
			continue
		}
		split := strings.Split(name, ",")
		if split[0] != "" && split[0] != "-" {
			names = append(names, split[0])
		}
	}

	return names
}
