package main

import (
	"fmt"
	"testing"
)

func testHelper(input string, expected int, t *testing.T) {
	result, err := calculate(input)
	if err != nil && expected != -1 {
		t.Error(fmt.Printf("Expected %d, got nil", expected))
	}
	if result != int64(expected) {
		t.Error(fmt.Printf("Expected %d, got %d", expected, result))
	}
}

func TestCalculate(t *testing.T) {
	var input string
	var expected int

	input = ""
	expected = -1
	testHelper(input, expected, t)

	input = "0"
	expected = 0
	testHelper(input, expected, t)

	input = "123"
	expected = 123
	testHelper(input, expected, t)

	input = "123k"
	expected = 123 * 1024
	testHelper(input, expected, t)

	input = "123M"
	expected = 123 * 1024 * 1024
	testHelper(input, expected, t)

	input = "123G"
	expected = 123 * 1024 * 1024 * 1024
	testHelper(input, expected, t)

	input = "1X"
	expected = -1
	testHelper(input, expected, t)

	input = "-1"
	expected = -1
	testHelper(input, expected, t)
}
