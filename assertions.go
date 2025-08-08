package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type AssertionEngine struct{}

func NewAssertionEngine() *AssertionEngine {
	return &AssertionEngine{}
}

func (ae *AssertionEngine) RunAssertions(result *TestResult, assertions []Assertion, variables map[string]string) []string {
	var errors []string

	for _, assertion := range assertions {
		if err := ae.runSingleAssertion(result, assertion, variables); err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors
}

func (ae *AssertionEngine) runSingleAssertion(result *TestResult, assertion Assertion, variables map[string]string) error {
	// Interpolate variables in assertion path and expected value
	interpolatedAssertion := ae.interpolateAssertion(assertion, variables)
	switch interpolatedAssertion.Type {
	case "status_code":
		return ae.assertStatusCode(result, interpolatedAssertion)
	case "json_path":
		return ae.assertJSONPath(result, interpolatedAssertion)
	case "xpath", "css_selector":
		return ae.assertHTMLSelector(result, interpolatedAssertion)
	case "header":
		return ae.assertHeader(result, interpolatedAssertion)
	case "body_contains":
		return ae.assertBodyContains(result, interpolatedAssertion)
	case "regex":
		return ae.assertRegex(result, interpolatedAssertion)
	case "response_time":
		return ae.assertResponseTime(result, interpolatedAssertion)
	default:
		return fmt.Errorf("unknown assertion type: %s", interpolatedAssertion.Type)
	}
}

func (ae *AssertionEngine) assertStatusCode(result *TestResult, assertion Assertion) error {
	expected, ok := assertion.Expected.(int)
	if !ok {
		if str, ok := assertion.Expected.(string); ok {
			var err error
			expected, err = strconv.Atoi(str)
			if err != nil {
				return fmt.Errorf("invalid status code format: %s", str)
			}
		} else {
			return fmt.Errorf("expected status code must be an integer")
		}
	}

	operator := assertion.Operator
	if operator == "" {
		operator = "equals"
	}

	switch operator {
	case "equals", "==":
		if result.StatusCode != expected {
			return fmt.Errorf("status code assertion failed: expected %d, got %d", expected, result.StatusCode)
		}
	case "not_equals", "!=":
		if result.StatusCode == expected {
			return fmt.Errorf("status code assertion failed: expected not %d, got %d", expected, result.StatusCode)
		}
	case "greater_than", ">":
		if result.StatusCode <= expected {
			return fmt.Errorf("status code assertion failed: expected > %d, got %d", expected, result.StatusCode)
		}
	case "less_than", "<":
		if result.StatusCode >= expected {
			return fmt.Errorf("status code assertion failed: expected < %d, got %d", expected, result.StatusCode)
		}
	default:
		return fmt.Errorf("unsupported operator for status code: %s", operator)
	}

	return nil
}

func (ae *AssertionEngine) assertJSONPath(result *TestResult, assertion Assertion) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(result.Response), &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	value, err := ae.getJSONPathValue(jsonData, assertion.Path)
	if err != nil {
		return fmt.Errorf("failed to extract JSON path %s: %w", assertion.Path, err)
	}

	operator := assertion.Operator
	if operator == "" {
		operator = "equals"
	}

	return ae.compareValues(value, assertion.Expected, operator, "JSON path")
}

func (ae *AssertionEngine) assertHTMLSelector(result *TestResult, assertion Assertion) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result.Response))
	if err != nil {
		return fmt.Errorf("failed to parse HTML response: %w", err)
	}

	var value interface{}
	selection := doc.Find(assertion.Path)

	if selection.Length() == 0 {
		value = nil
	} else if selection.Length() == 1 {
		value = strings.TrimSpace(selection.Text())
	} else {
		var values []string
		selection.Each(func(i int, s *goquery.Selection) {
			values = append(values, strings.TrimSpace(s.Text()))
		})
		value = values
	}

	operator := assertion.Operator
	if operator == "" {
		operator = "equals"
	}

	return ae.compareValues(value, assertion.Expected, operator, "HTML selector")
}

func (ae *AssertionEngine) assertHeader(result *TestResult, assertion Assertion) error {
	headerValues := result.Headers[assertion.Path]
	if len(headerValues) == 0 {
		return fmt.Errorf("header %s not found", assertion.Path)
	}

	var value interface{}
	if len(headerValues) == 1 {
		value = headerValues[0]
	} else {
		value = headerValues
	}

	operator := assertion.Operator
	if operator == "" {
		operator = "equals"
	}

	return ae.compareValues(value, assertion.Expected, operator, "header")
}

func (ae *AssertionEngine) assertBodyContains(result *TestResult, assertion Assertion) error {
	expected, ok := assertion.Expected.(string)
	if !ok {
		return fmt.Errorf("expected value for body_contains must be a string")
	}

	operator := assertion.Operator
	if operator == "" {
		operator = "contains"
	}

	switch operator {
	case "contains":
		if !strings.Contains(result.Response, expected) {
			return fmt.Errorf("body does not contain expected text: %s", expected)
		}
	case "not_contains":
		if strings.Contains(result.Response, expected) {
			return fmt.Errorf("body contains unexpected text: %s", expected)
		}
	default:
		return fmt.Errorf("unsupported operator for body_contains: %s", operator)
	}

	return nil
}

func (ae *AssertionEngine) assertRegex(result *TestResult, assertion Assertion) error {
	pattern, ok := assertion.Expected.(string)
	if !ok {
		return fmt.Errorf("expected value for regex must be a string pattern")
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	operator := assertion.Operator
	if operator == "" {
		operator = "matches"
	}

	switch operator {
	case "matches":
		if !regex.MatchString(result.Response) {
			return fmt.Errorf("response does not match regex pattern: %s", pattern)
		}
	case "not_matches":
		if regex.MatchString(result.Response) {
			return fmt.Errorf("response matches regex pattern (should not): %s", pattern)
		}
	default:
		return fmt.Errorf("unsupported operator for regex: %s", operator)
	}

	return nil
}

func (ae *AssertionEngine) assertResponseTime(result *TestResult, assertion Assertion) error {
	expectedMs, ok := assertion.Expected.(int)
	if !ok {
		if str, ok := assertion.Expected.(string); ok {
			var err error
			expectedMs, err = strconv.Atoi(str)
			if err != nil {
				return fmt.Errorf("invalid response time format: %s", str)
			}
		} else {
			return fmt.Errorf("expected response time must be an integer (milliseconds)")
		}
	}

	actualMs := int(result.Duration.Milliseconds())

	operator := assertion.Operator
	if operator == "" {
		operator = "less_than"
	}

	switch operator {
	case "less_than", "<":
		if actualMs >= expectedMs {
			return fmt.Errorf("response time assertion failed: expected < %dms, got %dms", expectedMs, actualMs)
		}
	case "greater_than", ">":
		if actualMs <= expectedMs {
			return fmt.Errorf("response time assertion failed: expected > %dms, got %dms", expectedMs, actualMs)
		}
	case "equals", "==":
		if actualMs != expectedMs {
			return fmt.Errorf("response time assertion failed: expected %dms, got %dms", expectedMs, actualMs)
		}
	default:
		return fmt.Errorf("unsupported operator for response_time: %s", operator)
	}

	return nil
}

func (ae *AssertionEngine) getJSONPathValue(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(strings.TrimPrefix(path, "$."), ".")
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			key := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", indexStr)
			}

			if key != "" {
				currentMap, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected object at path %s", key)
				}
				current = currentMap[key]
			}

			currentArray, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at index %d", index)
			}
			if index >= len(currentArray) {
				return nil, fmt.Errorf("array index %d out of bounds", index)
			}
			current = currentArray[index]
		} else {
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at path %s", part)
			}
			current = currentMap[part]
		}
	}

	return current, nil
}

func (ae *AssertionEngine) compareValues(actual, expected interface{}, operator, context string) error {
	// Normalize types before comparison
	normalizedActual, normalizedExpected := ae.normalizeTypes(actual, expected)
	switch operator {
	case "equals", "==":
		if !reflect.DeepEqual(normalizedActual, normalizedExpected) {
			return fmt.Errorf("%s assertion failed: expected %v, got %v", context, expected, actual)
		}
	case "not_equals", "!=":
		if reflect.DeepEqual(normalizedActual, normalizedExpected) {
			return fmt.Errorf("%s assertion failed: expected not %v, got %v", context, expected, actual)
		}
	case "contains":
		actualStr := fmt.Sprintf("%v", actual)
		expectedStr := fmt.Sprintf("%v", expected)
		if !strings.Contains(actualStr, expectedStr) {
			return fmt.Errorf("%s assertion failed: %v does not contain %v", context, actual, expected)
		}
	case "not_contains":
		actualStr := fmt.Sprintf("%v", actual)
		expectedStr := fmt.Sprintf("%v", expected)
		if strings.Contains(actualStr, expectedStr) {
			return fmt.Errorf("%s assertion failed: %v contains %v", context, actual, expected)
		}
	default:
		return fmt.Errorf("unsupported operator: %s", operator)
	}

	return nil
}

// normalizeTypes converts numeric types to ensure proper comparison
// JSON numbers are parsed as float64, but YAML might parse them as int
func (ae *AssertionEngine) normalizeTypes(actual, expected interface{}) (interface{}, interface{}) {
	// Handle numeric type conversions
	actualFloat, actualIsFloat := actual.(float64)
	expectedFloat, expectedIsFloat := expected.(float64)
	actualInt, actualIsInt := actual.(int)
	expectedInt, expectedIsInt := expected.(int)

	// If one is float64 and the other is int, convert both to float64
	if actualIsFloat && expectedIsInt {
		return actualFloat, float64(expectedInt)
	}
	if actualIsInt && expectedIsFloat {
		return float64(actualInt), expectedFloat
	}

	// If actual is float64 but represents a whole number, and expected is int
	if actualIsFloat && expectedIsInt {
		if actualFloat == float64(int(actualFloat)) {
			return actualFloat, float64(expectedInt)
		}
	}

	// If expected is float64 but represents a whole number, and actual is int
	if actualIsInt && expectedIsFloat {
		if expectedFloat == float64(int(expectedFloat)) {
			return float64(actualInt), expectedFloat
		}
	}

	// No conversion needed
	return actual, expected
}

// interpolateAssertion applies variable interpolation to assertion fields
func (ae *AssertionEngine) interpolateAssertion(assertion Assertion, variables map[string]string) Assertion {
	if variables == nil {
		return assertion
	}
	
	// Create a copy of the assertion to avoid modifying the original
	interpolated := Assertion{
		Type:     assertion.Type,
		Path:     interpolateVariables(assertion.Path, variables),
		Operator: assertion.Operator,
		Expected: ae.interpolateExpectedValue(assertion.Expected, variables),
	}
	
	return interpolated
}

// interpolateExpectedValue handles variable interpolation for different expected value types
func (ae *AssertionEngine) interpolateExpectedValue(expected interface{}, variables map[string]string) interface{} {
	if variables == nil {
		return expected
	}
	
	switch v := expected.(type) {
	case string:
		interpolated := interpolateVariables(v, variables)
		
		// If the original was a variable like "${user_id}" and it got interpolated to a number,
		// try to convert to the appropriate type
		if interpolated != v && strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			// Try integer conversion
			if intVal, err := strconv.Atoi(interpolated); err == nil {
				return intVal
			}
			// Try float conversion  
			if floatVal, err := strconv.ParseFloat(interpolated, 64); err == nil {
				return floatVal
			}
			// Try boolean conversion
			if boolVal, err := strconv.ParseBool(interpolated); err == nil {
				return boolVal
			}
		}
		
		return interpolated
	case int, float64, bool:
		// Non-string types are returned as-is since they can't contain variables
		return v
	default:
		// For complex types, try to convert to string, interpolate, then return
		if str := fmt.Sprintf("%v", expected); str != "" {
			interpolated := interpolateVariables(str, variables)
			if interpolated != str {
				return interpolated
			}
		}
		return expected
	}
}
