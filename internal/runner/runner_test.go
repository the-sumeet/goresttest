package runner

import (
	"goresttest/internal/config"
	"os"
	"testing"
)

type TestCases struct {
	testCase       config.TestCase
	expectedResult TestResult
}

func TestRunTest(t *testing.T) {

	// Set env vars
	os.Setenv("goresttest_example", "example.com")

	// Test cases
	testCases := []TestCases{
		{
			testCase: config.TestCase{
				Name:   "Test Case Success",
				URL:    "http://example.com",
				Method: "GET",
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Case Not Found Partial Status",
				URL:    "http://example.com/foo",
				Method: "GET",
				Validation: []config.Validation{
					{
						StatusCode: "4",
					},
				},
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Case Not Found",
				URL:    "http://example.com/foo",
				Method: "GET",
				Validation: []config.Validation{
					{StatusCode: "404"},
				},
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Case Not Found Partial Invalid Status ",
				URL:    "http://example.com/foo",
				Method: "GET",
				Validation: []config.Validation{
					{
						StatusCode: "42",
					},
				},
			},
			expectedResult: TestResult{
				Passed: false,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test templating",
				URL:    "http://{{ .goresttest_example }}",
				Method: "GET",
				Validation: []config.Validation{
					{
						StatusCode: "200",
					},
				},
			},
			expectedResult: TestResult{
				Passed: true,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Templating Wrong Variable",
				URL:    "http://{{ .goresttest_foo }}",
				Method: "GET",
				Validation: []config.Validation{
					{
						StatusCode: "200",
					},
				},
			},
			expectedResult: TestResult{
				Passed: false,
			},
		},
		{
			testCase: config.TestCase{
				Name:   "Test Validator",
				URL:    "https://jsonplaceholder.typicode.com/todos/1",
				Method: "GET",
				Validation: []config.Validation{
					{
						StatusCode: "200",
						Compare: config.Compare{
							JSONPath:   "userId",
							Comparator: "eq",
							Expected:   1,
						},
					},
				},
			},
			expectedResult: TestResult{
				Passed: true,
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
