package runner

import (
	"goresttest/internal/config"
	"testing"
)

type TestCases struct {
	testCase       config.TestCase
	expectedResult TestResult
}

func TestRunTest(t *testing.T) {

	// Test cases
	testCases := []TestCases{
		{
			testCase: config.TestCase{
				Name:   "Test Case 1",
				URL:    "http://example.com",
				Method: "GET",
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Case 1",
				URL:    "http://example.com/foo",
				Method: "GET",
				Validation: config.Validation{
					StatusCode: "4",
				},
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Case 1",
				URL:    "http://example.com/foo",
				Method: "GET",
				Validation: config.Validation{
					StatusCode: "404",
				},
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Case 1",
				URL:    "http://example.com/foo",
				Method: "GET",
				Validation: config.Validation{
					StatusCode: "42",
				},
			},
			expectedResult: TestResult{
				Passed: false,
			},
		},
	}

	for _, tc := range testCases {
		result := RunTest(tc.testCase)
		if result.Passed != tc.expectedResult.Passed {
			t.Errorf("Expected %v, got %v", tc.expectedResult, result)
		}
	}
}
