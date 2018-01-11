package errors

import (
	"reflect"
	"sort"
)

func deepAppendTypes(types []string, err error) []string {
	walk(err, func(err error) {
		types = appendTypes(types, err)
	})
	return dedupeTypes(types)
}

func appendTypes(types []string, err error) []string {
	if e, ok := err.(errorTypes); ok {
		types = append(types, e.Types()...)
	}

	t := reflect.TypeOf(err)
	v := reflect.ValueOf(err)

	for i, n := 0, t.NumMethod(); i != n; i++ {
		mt := t.Method(i)
		mv := v.Method(i)

		if f, ok := mv.Interface().(func() bool); ok && f() {
			types = append(types, mt.Name)
		}
	}

	return types
}

func copyTypes(types []string) []string {
	if len(types) == 0 {
		return nil
	}
	cpy := make([]string, len(types))
	copy(cpy, types)
	return cpy
}

func dedupeTypes(types []string) []string {
	if len(types) == 0 {
		return nil
	}

	sortTypes(types)

	prev := types[0]
	i := 1
	j := 1
	n := len(types)

	for i != n {
		if types[i] != prev {
			prev, types[j] = types[i], types[i]
			j++
		}
		i++
	}

	return types[:j]
}

func sortTypes(types []string) {
	sort.Strings(types)
}
