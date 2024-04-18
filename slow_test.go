package main

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

// func BenchmarkPipeline1(b *testing.B) {
// 	var rChan <-chan string

// 	for i := 0; i < b.N; i++ {
// 		rChan = RunPipeline1(context.Background(), source)
// 	}

// 	resChan = rChan

// }

func TestSlowPipeline(t *testing.T) {
	source := []string{"ANOTHERLONGSTRING", "YETANOTHERONE", "FOO", "", "BAR"}

	expected := []string{"Anotherlongstring", "Yetanotherone", "Foo", "", "Bar"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actualChan := RunPipeline1(ctx, source)
	actual := make([]string, 0)
	for val := range actualChan {
		actual = append(actual, val)
	}

	sort.Strings(expected)
	sort.Strings(actual)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
