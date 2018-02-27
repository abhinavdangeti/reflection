package reflection

import "reflect"

type YY struct {
	a uint64
	b uint64
}

type WW struct {
	a int
	b uint64
	c float64
	d string
	e *YY
}

type X struct {
	name   string
	arr    []*WW
	theMap map[string]*WW
	num    uint64
}

var xSizeStatic int
var xSizeDynamic []dynamicSizeFunc

func init() {
	var x X
	Register("size.X", x)
}

func (x X) Size() int {
	return SizeOf("size.X", x)
}

func (x X) SizeReflect() int {
	xT := reflect.TypeOf(x)
	rv := int(xT.Size())

	xV := reflect.ValueOf(x)
	rv += getSizeViaReflection(xV, rv)

	return rv
}

func getSizeViaReflection(x reflect.Value, rv int) int {
	switch x.Kind() {
		case reflect.Struct:
			for i := 0; i < x.NumField(); i++ {
				rv = getSizeViaReflection(x.Field(i), rv)
			}
		case reflect.Ptr:
			rv += 8
			rv = getSizeViaReflection(x.Elem(), rv)
		case reflect.Slice:
			rv += 24
			for i := 0; i < x.Len(); i++ {
				rv = getSizeViaReflection(x.Index(i), rv)
			}
		case reflect.Map:
			rv += 8
			keys := x.MapKeys()
			for _, entry := range keys {
				rv = getSizeViaReflection(entry, rv)
				rv = getSizeViaReflection(x.MapIndex(entry), rv)
			}
		case reflect.String:
			rv += 16 + x.Len()
		case reflect.Uint64, reflect.Int, reflect.Float64:
			rv += 8
	}
	return rv
}

func (x X) SizeManual() int {
	sizeInBytes := 8 // pointer

	sizeInBytes += 16 + len(x.name) +
		24 +
		8 +
		8

	for _, entry := range x.arr {
		sizeInBytes += 8 + // pointer
			8 +
			8 +
			8 +
			16 + len(entry.d) +
			8 + 8*2
	}

	for k, v := range x.theMap {
		sizeInBytes += 16 + len(k) +
			8 + 8*3 + 16 + len(v.d) + 8 + 8*2
	}

	return sizeInBytes
}
