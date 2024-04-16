package main

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func BenchmarkSlowPipeline(b *testing.B) {

	for i := 0; i < b.N; i++ {
		RunSlowPipeline(context.Background(), source)
	}

}

func TestSlowPipeline(t *testing.T) {
	source := []string{"ANOTHERLONGSTRING", "YETANOTHERONE", "FOO", "", "BAR"}

	expected := []string{"Anotherlongstring", "Yetanotherone", "Foo", "", "Bar"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actualChan := RunSlowPipeline(ctx, source)
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
