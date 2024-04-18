package main

import (
	"context"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

func generateLongString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func generateStringSlice(length, count int) []string {
	var result []string
	for i := 0; i < count; i++ {
		result = append(result, generateLongString(length))
	}
	return result
}

var source = generateStringSlice(30, 10)

var resChan <-chan string

func BenchmarkPipeline2(b *testing.B) {

	var rChan <-chan string
	for i := 0; i < b.N; i++ {
		rChan = RunPipeline2(context.Background(), source)
	}

	resChan = rChan

}

func TestFastPipeline(t *testing.T) {
	source := []string{"FOO", "BAR", "BAX", "", "XYZ"}

	expected := []string{"Foo", "Bar", "Bax", "", "Xyz"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actualChan := RunPipeline2(ctx, source)
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
