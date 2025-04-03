package config

import (
	"reflect"
	"testing"
)

func TestLoadTestCases(t *testing.T) {

	expectedData := []TestCase{
		{
			Name:   "Get public API",
			URL:    "https://jsonplaceholder.typicode.com",
			Method: "GET",
		},
		{
			Name:   "Test invalid endpoint",
			URL:    "https://jsonplaceholder.typicode.com/foo",
			Method: "GET",
			Validation: Validation{
				StatusCode: "4",
			},
		},
		{
			Name:   "Test invalid endpoint",
			URL:    "https://jsonplaceholder.typicode.com/foo",
			Method: "GET",
			Validation: Validation{
				StatusCode: "404",
			},
		},
	}

	testData, err := LoadTestCases("test.yml")
	if err != nil {
		t.Fatalf("Failed to load test cases: %v", err)
	}

	if !reflect.DeepEqual(testData, expectedData) {
		t.Fatalf("Expected %v, got %v", expectedData, testData)
	}

}
