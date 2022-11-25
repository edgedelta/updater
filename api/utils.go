package api

import (
	"fmt"
	"strings"

	"github.com/fatih/structs"
)

func SetStructFieldValue(o any, path []string, setValue any) error {
	if len(path) == 0 {
		return fmt.Errorf("no path specified")
	}

	fields := structs.Fields(o)
	lookForTag := path[0]

	for _, f := range fields {
		sp := strings.Split(f.Tag("json"), ",")
		tag := sp[0]
		if tag != lookForTag {
			continue
		}
		if len(path) == 1 {
			return f.Set(setValue)
		}
		newPath := path[1:]
		return SetStructFieldValue(f.Value(), newPath, setValue)
	}

	return fmt.Errorf("could not find field with JSON tag %s in object %+v", lookForTag, o)
}
