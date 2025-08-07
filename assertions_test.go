package main

import (
	"testing"
	"time"
)

func TestAssertionEngine_StatusCode(t *testing.T) {
	engine := NewAssertionEngine()

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "status code equals - success",
			result: &TestResult{
				StatusCode: 200,
			},
			assertion: Assertion{
				Type:     "status_code",
				Expected: 200,
				Operator: "equals",
			},
			wantError: false,
		},
		{
			name: "status code equals - failure",
			result: &TestResult{
				StatusCode: 404,
			},
			assertion: Assertion{
				Type:     "status_code",
				Expected: 200,
			},
			wantError: true,
		},
		{
			name: "status code not equals - success",
			result: &TestResult{
				StatusCode: 404,
			},
			assertion: Assertion{
				Type:     "status_code",
				Expected: 200,
				Operator: "not_equals",
			},
			wantError: false,
		},
		{
			name: "status code greater than - success",
			result: &TestResult{
				StatusCode: 201,
			},
			assertion: Assertion{
				Type:     "status_code",
				Expected: 200,
				Operator: "greater_than",
			},
			wantError: false,
		},
		{
			name: "status code less than - success",
			result: &TestResult{
				StatusCode: 199,
			},
			assertion: Assertion{
				Type:     "status_code",
				Expected: 200,
				Operator: "less_than",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_JSONPath(t *testing.T) {
	engine := NewAssertionEngine()

	jsonResponse := `{
		"name": "John Doe",
		"age": 30,
		"email": "john@example.com",
		"active": true,
		"scores": [85, 90, 78],
		"address": {
			"city": "New York",
			"zipcode": "10001"
		}
	}`

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "json path string equals - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "name",
				Expected: "John Doe",
			},
			wantError: false,
		},
		{
			name: "json path number equals - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "age",
				Expected: float64(30),
			},
			wantError: false,
		},
		{
			name: "json path boolean equals - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "active",
				Expected: true,
			},
			wantError: false,
		},
		{
			name: "json path array element - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "scores[0]",
				Expected: float64(85),
			},
			wantError: false,
		},
		{
			name: "json path nested object - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "address.city",
				Expected: "New York",
			},
			wantError: false,
		},
		{
			name: "json path string not equals - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "name",
				Expected: "Jane Doe",
				Operator: "not_equals",
			},
			wantError: false,
		},
		{
			name: "json path equals - failure",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "name",
				Expected: "Jane Doe",
			},
			wantError: true,
		},
		{
			name: "json path invalid - failure",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "nonexistent",
				Expected: "anything",
			},
			wantError: true,
		},
		{
			name: "json path number with int expected - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "age",
				Expected: 30, // int instead of float64
			},
			wantError: false,
		},
		{
			name: "json path array element with int expected - success",
			result: &TestResult{
				Response: jsonResponse,
			},
			assertion: Assertion{
				Type:     "json_path",
				Path:     "scores[0]",
				Expected: 85, // int instead of float64
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_HTMLSelector(t *testing.T) {
	engine := NewAssertionEngine()

	htmlResponse := `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
</head>
<body>
	<h1 id="main-title">Welcome</h1>
	<div class="content">
		<p>Hello World</p>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
			<li>Item 3</li>
		</ul>
	</div>
	<div class="footer">
		<span class="copyright">© 2023</span>
	</div>
</body>
</html>`

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "css selector title - success",
			result: &TestResult{
				Response: htmlResponse,
			},
			assertion: Assertion{
				Type:     "css_selector",
				Path:     "title",
				Expected: "Test Page",
			},
			wantError: false,
		},
		{
			name: "css selector h1 - success",
			result: &TestResult{
				Response: htmlResponse,
			},
			assertion: Assertion{
				Type:     "css_selector",
				Path:     "#main-title",
				Expected: "Welcome",
			},
			wantError: false,
		},
		{
			name: "css selector p - success",
			result: &TestResult{
				Response: htmlResponse,
			},
			assertion: Assertion{
				Type:     "css_selector",
				Path:     ".content p",
				Expected: "Hello World",
			},
			wantError: false,
		},
		{
			name: "css selector first li - success",
			result: &TestResult{
				Response: htmlResponse,
			},
			assertion: Assertion{
				Type:     "css_selector",
				Path:     "li:first-child",
				Expected: "Item 1",
			},
			wantError: false,
		},
		{
			name: "css selector nonexistent - failure",
			result: &TestResult{
				Response: htmlResponse,
			},
			assertion: Assertion{
				Type:     "css_selector",
				Path:     ".nonexistent",
				Expected: "anything",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_Header(t *testing.T) {
	engine := NewAssertionEngine()

	headers := map[string][]string{
		"Content-Type":   {"application/json"},
		"Content-Length": {"123"},
		"Set-Cookie":     {"session=abc", "user=xyz"},
	}

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "header single value - success",
			result: &TestResult{
				Headers: headers,
			},
			assertion: Assertion{
				Type:     "header",
				Path:     "Content-Type",
				Expected: "application/json",
			},
			wantError: false,
		},
		{
			name: "header missing - failure",
			result: &TestResult{
				Headers: headers,
			},
			assertion: Assertion{
				Type:     "header",
				Path:     "Authorization",
				Expected: "Bearer token",
			},
			wantError: true,
		},
		{
			name: "header wrong value - failure",
			result: &TestResult{
				Headers: headers,
			},
			assertion: Assertion{
				Type:     "header",
				Path:     "Content-Type",
				Expected: "text/html",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_BodyContains(t *testing.T) {
	engine := NewAssertionEngine()

	response := "This is a test response containing some specific text"

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "body contains - success",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "body_contains",
				Expected: "specific text",
			},
			wantError: false,
		},
		{
			name: "body not contains - success",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "body_contains",
				Expected: "missing text",
				Operator: "not_contains",
			},
			wantError: false,
		},
		{
			name: "body contains - failure",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "body_contains",
				Expected: "missing text",
			},
			wantError: true,
		},
		{
			name: "body not contains - failure",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "body_contains",
				Expected: "specific text",
				Operator: "not_contains",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_Regex(t *testing.T) {
	engine := NewAssertionEngine()

	response := "User ID: 12345, Email: user@example.com"

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "regex matches - success",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "regex",
				Expected: `User ID: \d+`,
			},
			wantError: false,
		},
		{
			name: "regex email matches - success",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "regex",
				Expected: `\w+@\w+\.\w+`,
			},
			wantError: false,
		},
		{
			name: "regex not matches - success",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "regex",
				Expected: `Phone: \d+`,
				Operator: "not_matches",
			},
			wantError: false,
		},
		{
			name: "regex matches - failure",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "regex",
				Expected: `Phone: \d+`,
			},
			wantError: true,
		},
		{
			name: "regex not matches - failure",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "regex",
				Expected: `User ID: \d+`,
				Operator: "not_matches",
			},
			wantError: true,
		},
		{
			name: "invalid regex - failure",
			result: &TestResult{
				Response: response,
			},
			assertion: Assertion{
				Type:     "regex",
				Expected: `[invalid`,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_ResponseTime(t *testing.T) {
	engine := NewAssertionEngine()

	tests := []struct {
		name      string
		result    *TestResult
		assertion Assertion
		wantError bool
	}{
		{
			name: "response time less than - success",
			result: &TestResult{
				Duration: 50 * time.Millisecond,
			},
			assertion: Assertion{
				Type:     "response_time",
				Expected: 100,
				Operator: "less_than",
			},
			wantError: false,
		},
		{
			name: "response time greater than - success",
			result: &TestResult{
				Duration: 150 * time.Millisecond,
			},
			assertion: Assertion{
				Type:     "response_time",
				Expected: 100,
				Operator: "greater_than",
			},
			wantError: false,
		},
		{
			name: "response time equals - success",
			result: &TestResult{
				Duration: 100 * time.Millisecond,
			},
			assertion: Assertion{
				Type:     "response_time",
				Expected: 100,
				Operator: "equals",
			},
			wantError: false,
		},
		{
			name: "response time less than - failure",
			result: &TestResult{
				Duration: 150 * time.Millisecond,
			},
			assertion: Assertion{
				Type:     "response_time",
				Expected: 100,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.runSingleAssertion(tt.result, tt.assertion)
			if (err != nil) != tt.wantError {
				t.Errorf("runSingleAssertion() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAssertionEngine_RunAssertions(t *testing.T) {
	engine := NewAssertionEngine()

	result := &TestResult{
		StatusCode: 200,
		Response:   `{"name": "John", "age": 30}`,
		Duration:   50 * time.Millisecond,
	}

	assertions := []Assertion{
		{
			Type:     "status_code",
			Expected: 200,
		},
		{
			Type:     "json_path",
			Path:     "name",
			Expected: "John",
		},
		{
			Type:     "response_time",
			Expected: 100,
		},
	}

	errors := engine.RunAssertions(result, assertions)
	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	failingAssertions := []Assertion{
		{
			Type:     "status_code",
			Expected: 404,
		},
		{
			Type:     "json_path",
			Path:     "name",
			Expected: "Jane",
		},
	}

	errors = engine.RunAssertions(result, failingAssertions)
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d: %v", len(errors), errors)
	}
}
