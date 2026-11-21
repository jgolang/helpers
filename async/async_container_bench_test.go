package async

import (
	"strconv"
	"testing"
)

type benchFlatStruct struct {
	ID    int
	Name  string
	Score float64
}

type benchNestedStruct struct {
	Name   string
	Tags   []string
	Attrs  map[string]int
	Parent *benchFlatStruct
}

func BenchmarkContainerGet(b *testing.B) {
	attrs := make(map[string]int, 100)
	for i := 0; i < 100; i++ {
		attrs[strconv.Itoa(i)] = i
	}

	b.Run("int", func(b *testing.B) {
		benchmarkGet[int](b, 42)
	})
	b.Run("flat_struct", func(b *testing.B) {
		benchmarkGet[benchFlatStruct](b, benchFlatStruct{ID: 1, Name: "alice", Score: 9.5})
	})
	b.Run("slice_1k_int", func(b *testing.B) {
		benchmarkGet[[]int](b, make([]int, 1000))
	})
	b.Run("map_100", func(b *testing.B) {
		benchmarkGet[map[string]int](b, attrs)
	})
	b.Run("nested_struct", func(b *testing.B) {
		benchmarkGet[benchNestedStruct](b, benchNestedStruct{
			Name:   "node",
			Tags:   []string{"a", "b", "c"},
			Attrs:  map[string]int{"x": 1, "y": 2},
			Parent: &benchFlatStruct{ID: 1, Name: "root"},
		})
	})
}

func benchmarkGet[T any](b *testing.B, stored T) {
	container := &Container{}
	container.Set("key", stored, nil)
	var got T

	b.ReportAllocs()
	for b.Loop() {
		if err := container.Get("key", &got); err != nil {
			b.Fatal(err)
		}
	}
}
