package runner

import (
	"bytes"
	"errors"
	"goresttest/internal/config"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type TestResult struct {
	TestCase   config.TestCase
	Passed     bool
	StatusCode int
	Error      error
}

func RunTest(testCase config.TestCase) (result TestResult) {

	data := map[string]interface{}{}
	getEnvVars(data)

	result = TestResult{Passed: false, TestCase: testCase}

	client := &http.Client{}

	testCase, err := interpolateVariables(testCase, data)
	if err != nil {
		result.Error = err
		return
	}

	req, err := http.NewRequest(testCase.Method, testCase.URL, bytes.NewBufferString(testCase.Body))
	if err != nil {
		result.Error = err
		return
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

	err = performValidation(testCase, resp)
	if err != nil {
		result.Error = err
		return result
	}

	// // Validate statuc code
	// if testCase.Validation.StatusCode == "" {
	// 	testCase.Validation.StatusCode = "2"
	// }

	// if !strings.HasPrefix(strconv.Itoa(resp.StatusCode), testCase.Validation.StatusCode) {
	// 	return
	// }

	result.Passed = true
	return result
}

func performValidation(testCase config.TestCase, resp *http.Response) error {

	for _, validation := range testCase.Validation {

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		// Validate status code
		statusCodeToCheck := validation.StatusCode
		if statusCodeToCheck == "" {
			statusCodeToCheck = "2"
		}

		if !strings.HasPrefix(strconv.Itoa(resp.StatusCode), statusCodeToCheck) {
			return errors.New("status code does not match")
		}

		// Validate compare
		jsonPath := validation.Compare.JSONPath
		if jsonPath != "" {
			value := gjson.Get(bodyString, jsonPath)
			if !value.Exists() {
				return errors.New("json path does not exist")
			}

			val := value.Value()
			expceted := validation.Compare.Expected

			if value.Type == gjson.Number {
				expcetedInt, ok := expceted.(int)
				if ok {
					expceted = float64(expcetedInt)
				}
			}

			if val != expceted {
				return errors.New("json path does not match")
			}
		}

	}

	return nil
}

func interpolateVariables(testCase config.TestCase, data map[string]interface{}) (config.TestCase, error) {
	t := template.Must(template.New("doesntmatter").Parse(testCase.URL))
	var finalUrl bytes.Buffer
	err := t.Execute(&finalUrl, data)
	if err != nil {
		return testCase, err
	}
	testCase.URL = finalUrl.String()
	return testCase, nil
}

func getEnvVars(data map[string]interface{}) {
	for _, env := range os.Environ() {
		keyValue := strings.SplitN(env, "=", 2)
		if len(keyValue) == 2 {
			data[keyValue[0]] = keyValue[1]
		}
	}
}
