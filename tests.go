package main

import (
	"reflect"
	"testing"
)

func TestSimpleFilterByPrefix_0(t *testing.T) {
	ts := generateTransactionsForTest(0)

	ptree := constructPrefixTree(ts)

	actual := filterByPrefixTree(ptree, "Monitor")
	expected := simpleFilterByPrefix(ts, "Monitor")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestSimpleFilterByPrefix_100(t *testing.T) {
	ts := generateTransactionsForTest(100)

	ptree := constructPrefixTree(ts)

	actual := filterByPrefixTree(ptree, "Monitor")
	expected := simpleFilterByPrefix(ts, "Monitor")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
