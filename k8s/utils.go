package k8s

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/structs"
)

var (
	sliceRe = regexp.MustCompile(`([^\[\]]+)\[(\d+)\]$`)
)

func SetStructFieldValue(o any, path []string, setValue any) error {
	if len(path) == 0 {
		return errors.New("no path specified")
	}
	fields := structs.Fields(o)
	lookForTag := path[0]
	wantSlice := false
	var sliceIndex uint64
	if match := sliceRe.FindStringSubmatch(lookForTag); match != nil {
		var err error
		sliceIndex, err = strconv.ParseUint(match[2], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse slice index '%s' to uint64, err: %v", match[2], err)
		}
		wantSlice = true
		lookForTag = match[1]
	}
	for _, f := range fields {
		sp := strings.Split(f.Tag("json"), ",")
		tag := sp[0]
		if tag != lookForTag {
			continue
		}
		if len(path) == 1 {
			if wantSlice {
				return errors.New("directly setting a slice element is not supported")
			}
			return f.Set(setValue)
		}
		var obj any = f.Value()
		if wantSlice {
			if f.Kind() != reflect.Slice {
				return fmt.Errorf("expected '%s' to be a slice, got %s instead", lookForTag, f.Kind().String())
			}
			obj = reflect.ValueOf(f.Value()).Index(int(sliceIndex)).Addr().Interface()
		}
		newPath := path[1:]
		return SetStructFieldValue(obj, newPath, setValue)
	}
	return fmt.Errorf("could not find field with JSON tag %s in object %+v", lookForTag, o)
}
