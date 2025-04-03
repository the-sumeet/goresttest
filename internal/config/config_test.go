package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadTestCases(t *testing.T) {

	testData, err := LoadTestCases("test.yml")
	if err != nil {
		t.Fatalf("Failed to load test cases: %v", err)
	}

	if len(testData) != 2 {
		t.Fatalf("Expected 2 test cases, got %d", len(testData))
	}

	testDataSerialized, err := yaml.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to serialize test cases: %v", err)
	}

	if string(testDataSerialized) != "- name: Get public API\n  url: https://api.publicapis.org/entries\n  method: GET\n  headers: {}\n  body: \"\"\n- name: Test invalid endpoint\n  url: https://api.publicapis.org/invalid\n  method: GET\n  headers: {}\n  body: \"\"\n" {
		t.Fatalf("Serialized test cases do not match expected output")
	}

}
