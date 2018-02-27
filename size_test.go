package reflection

import (
	"fmt"
	"strconv"
	"testing"
)

var xSize int
var ySize int
var zSize int

var xx *X

func init() {
	xx = &X{
		name:   "hello",
		arr:    make([]*WW, 10),
		theMap: map[string]*WW{},
		num:    10000,
	}

	for i := 0; i < 10; i++ {
		iStr := strconv.Itoa(int(i * 10000))
		w := &WW{
			a: int(i * 10),
			b: uint64(i * 100),
			c: float64(i) * 1.1,
			d: iStr,
			e: &YY{a: uint64(i), b: uint64(i + 5)},
		}

		xx.arr[i] = w
		xx.theMap[iStr] = w
	}
}

func BenchmarkSize(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		xSize = xx.Size()
	}
}

func BenchmarkSizeReflect(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ySize = xx.SizeReflect()
	}
}

func BenchmarkSizeManual(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		zSize = xx.SizeManual()
	}
}

func TestBasic(t *testing.T) {
	fmt.Printf("REFLECT INIT'ED: %v\n", xx.Size())
	fmt.Printf("REFLECT RUNTIME: %v\n", xx.SizeReflect())
	fmt.Printf("MANUAL CALC: %v\n", xx.SizeManual())
}
