package runner

import (
	"bytes"
	"goresttest/internal/config"
	"net/http"
	"strconv"
	"strings"
)

type TestResult struct {
	TestCase   config.TestCase
	Passed     bool
	StatusCode int
	Error      error
}

func RunTest(testCase config.TestCase) TestResult {
	result := TestResult{Passed: true, TestCase: testCase}

	client := &http.Client{}
	req, err := http.NewRequest(testCase.Method, testCase.URL, bytes.NewBufferString(testCase.Body))
	if err != nil {
		result.Error = err
		return result
	}

	// Set headers
	for key, value := range testCase.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Validate statuc code
	if testCase.Validation.StatusCode == "" {
		testCase.Validation.StatusCode = "2"
	}

	if !strings.HasPrefix(strconv.Itoa(resp.StatusCode), testCase.Validation.StatusCode) {
		result.Passed = false
	}

	return result
}
