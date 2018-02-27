package reflection

import (
//	"fmt"
	"reflect"
)

var registryStatic map[string]int
var registryDynamic map[string][]dynamicSizeFunc

func init() {
	registryStatic = make(map[string]int)
	registryDynamic = make(map[string][]dynamicSizeFunc)
}

func Register(name string, i interface{}) {
	iT := reflect.TypeOf(i)
	if iT.Kind() == reflect.Ptr {
		registryStatic[name] = int(iT.Size() + iT.Elem().Size())
		registryDynamic[name] = buildDynamic(iT.Elem(), nil, 0)
	} else {
		registryStatic[name] = int(iT.Size())
		registryDynamic[name] = buildDynamic(iT, nil, 0)
	}
}

func SizeOf(name string, i interface{}) int {
	iV := reflect.ValueOf(i)
	rv := registryStatic[name]
	ids := registryDynamic[name]

	rvDynamic, _ := recur(ids, -1, iV, rv, true)
	rv += rvDynamic

	return rv
}

// returns rv (size estimate) and index (cursor position in ids)
func recur(ids []dynamicSizeFunc, index int, val reflect.Value, rv int, loop bool) (int, int) {
	index++
	if index >= len(ids) {
		return rv, index
	}

	if val.Kind() == reflect.Ptr {
		rv += 8
		val = val.Elem()
	}

	curr := ids[index](val)
	rv += curr.size
	if curr.count > 0 {
		// case of slices or maps
		if curr.value.Kind() == reflect.Slice {
			var num_fields int
			for i := 0; i < curr.count; i++ {
				entry := curr.value.Index(i)
				rv, num_fields = processEntry(ids, index, entry, rv)
			}
			index += num_fields
		} else if curr.value.Kind() == reflect.Map {
			var key_fields, val_fields int
			keys := curr.value.MapKeys()
			for _, entry := range keys {
				rv, key_fields = processEntry(ids, index, entry, rv)
				rv, val_fields = processEntry(ids, index+key_fields, curr.value.MapIndex(entry), rv)
			}
			index += key_fields + val_fields
		}
	}

	if loop {
		return recur(ids, index, val, rv, loop)
	} else {
		return rv, index
	}
}

func processEntry(ids []dynamicSizeFunc, index int, entry reflect.Value, rv int) (int, int) {
	num_fields := 1
	if entry.Kind() == reflect.Ptr {
		entry = entry.Elem()
		rv += 8
	}
	if entry.Kind() == reflect.Struct {
		num_fields = entry.NumField()
	}
	for j := 0; j < num_fields; j++ {
		rv, _ = recur(ids, index+j, entry, rv, false)
	}

	return rv, num_fields
}

type status struct {
	size  int
	count int
	value reflect.Value
}

type dynamicSizeFunc func(reflect.Value) status

func buildDynamicForField(f dynamicSizeFunc, i int) dynamicSizeFunc {
	return func(v reflect.Value) status {
		if v.Kind() == reflect.Struct {
			return f(v.Field(i))
		} else {
			return f(v)
		}
	}
}

func buildDynamic(t reflect.Type, rv []dynamicSizeFunc, index int) []dynamicSizeFunc {
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			rv = buildDynamic(f.Type, rv, i)
		}
	case reflect.Ptr:
		rv = buildDynamic(t.Elem(), rv, index)
	case reflect.Slice:
		rv = append(rv, buildDynamicForField(dynamicSizeSlice, index))
//		fmt.Println(t.Kind().String(), " : ", index)
		rv = buildDynamic(t.Elem(), rv, index)
	case reflect.Map:
		rv = append(rv, buildDynamicForField(dynamicSizeMap, index))
//		fmt.Println(t.Kind().String(), " : ", index)
		rv = buildDynamic(t.Key(), rv, index)
		rv = buildDynamic(t.Elem(), rv, index)
	case reflect.String:
		rv = append(rv, buildDynamicForField(dynamicSizeString, index))
//		fmt.Println(t.Kind().String(), " : ", index)
	case reflect.Uint64, reflect.Int, reflect.Float64:
		rv = append(rv, buildDynamicForField(dynamicSizeNumber, index))
//		fmt.Println(t.Kind().String(), " : ", index)
	}

	return rv
}

func dynamicSizeSlice(v reflect.Value) status {
	return status{
		size:  24,
		count: v.Len(),
		value: v,
	}
}

func dynamicSizeMap(v reflect.Value) status {
	return status{
		size:  8,
		count: v.Len(),
		value: v,
	}
}

func dynamicSizeString(v reflect.Value) status {
	return status{
		size: 16 + v.Len(),
	}
}

func dynamicSizeNumber(v reflect.Value) status {
	return status{
		size: 8,
	}
}
