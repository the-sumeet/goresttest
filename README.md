# GoRestTest

GoRestTest is a powerful REST API testing framework for Go that can be used both as a library and as a CLI tool. It's inspired by pyresttest and provides comprehensive features for testing REST APIs including assertions, variable extraction, test dependencies, and parallel execution.

## Features

- **Dual Usage**: Can be used both as a Go library and as a CLI tool
- **Comprehensive Assertions**: Status codes, JSON paths, HTML selectors, headers, response time, regex matching, and body content
- **Variable Extraction**: Extract data from responses for use in subsequent tests
- **Test Dependencies**: Define test execution order with depends_on
- **Parallel Execution**: Run independent tests in parallel for faster execution
- **Multiple Report Formats**: Console, JSON, and HTML reports
- **Variable Interpolation**: Use variables in test definitions
- **Body File Support**: Load request bodies from external files

## Installation

### As a CLI Tool

```bash
go install github.com/the-sumeet/goresttest/cmd/goresttest@latest
```

### As a Library

```bash
go get github.com/the-sumeet/goresttest@latest
```

## Usage

### CLI Tool

```bash
# Basic usage
goresttest -config tests.yaml

# Parallel execution
goresttest -config tests.yaml -parallel -workers 5

# Generate HTML report
goresttest -config tests.yaml -output html -file report.html

# Generate JSON report
goresttest -config tests.yaml -output json -file report.json

# Verbose output
goresttest -config tests.yaml -verbose
```

### Library Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/the-sumeet/goresttest"
)

func main() {
    // Parse test suite from file
    suite, err := goresttest.ParseTestSuite("tests.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create test runner
    runner := goresttest.NewTestRunner(suite.BaseURL)
    
    // Execute test suite
    results, err := runner.RunTestSuite(suite)
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate report
    reporter := goresttest.NewReporter()
    reporter.PrintConsoleReport(results)
}
```

### Test Configuration Format

```yaml
name: "API Test Suite"
base_url: "https://api.example.com"
variables:
  user_id: "123"
  api_key: "your-api-key"
parallel: true
max_workers: 5

tests:
  - name: "Get User"
    method: "GET"
    url: "/users/${user_id}"
    headers:
      Authorization: "Bearer ${api_key}"
    assertions:
      - type: "status_code"
        expected: 200
      - type: "json_path"
        path: "$.id"
        expected: 123
      - type: "response_time"
        expected: 1000
        operator: "less_than"
    extract:
      username: "json:$.username"
  
  - name: "Update User"
    method: "PUT"
    url: "/users/${user_id}"
    headers:
      Authorization: "Bearer ${api_key}"
      Content-Type: "application/json"
    body: '{"name": "Updated Name"}'
    assertions:
      - type: "status_code"
        expected: 200
      - type: "json_path"
        path: "$.name"
        expected: "Updated Name"
```

## Assertion Types

### Status Code
```yaml
- type: "status_code"
  expected: 200
  operator: "equals"  # equals, not_equals, greater_than, less_than
```

### JSON Path
```yaml
- type: "json_path"
  path: "$.data.users[0].name"
  expected: "John Doe"
  operator: "equals"  # equals, not_equals, contains, not_contains
```

### HTML Selector
```yaml
- type: "css_selector"
  path: "h1.title"
  expected: "Welcome"
```

### Header
```yaml
- type: "header"
  path: "Content-Type"
  expected: "application/json"
```

### Body Contains
```yaml
- type: "body_contains"
  expected: "success"
  operator: "contains"  # contains, not_contains
```

### Regex
```yaml
- type: "regex"
  expected: "\\d{3}-\\d{2}-\\d{4}"
  operator: "matches"  # matches, not_matches
```

### Response Time
```yaml
- type: "response_time"
  expected: 1000
  operator: "less_than"  # less_than, greater_than, equals
```

## Variable Extraction

Extract data from responses for use in subsequent tests:

```yaml
extract:
  user_id: "json:$.id"                    # Extract from JSON response
  session_token: "header:X-Session-Token"  # Extract from response header
  csrf_token: "regex:<input name=\"_token\" value=\"([^\"]+)\""  # Extract using regex
  title: "css:h1.title"                    # Extract using CSS selector
  status_code: "status:"                   # Extract status code
  response_time: "response_time:"          # Extract response time
```

## Test Dependencies

Define test execution order:

```yaml
tests:
  - name: "Login"
    method: "POST"
    url: "/auth/login"
    # ... test definition

  - name: "Get Profile"
    method: "GET"
    url: "/profile"
    depends_on:
      - "Login"
    # ... test definition
```

## Programmatic Usage

### Creating Tests Programmatically

```go
test := goresttest.Test{
    Name:   "Create User",
    Method: "POST",
    URL:    "/users",
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
    Body: `{"name": "John Doe", "email": "john@example.com"}`,
    Assertions: []goresttest.Assertion{
        {
            Type:     "status_code",
            Expected: 201,
        },
        {
            Type:     "json_path",
            Path:     "$.name",
            Expected: "John Doe",
        },
    },
    Extract: map[string]string{
        "user_id": "json:$.id",
    },
}

runner := goresttest.NewTestRunner("https://api.example.com")
result, err := runner.RunTest(test, nil)
```

### Using Test Results

```go
if result.Success {
    fmt.Printf("Test passed in %v\n", result.Duration)
    fmt.Printf("Status Code: %d\n", result.StatusCode)
    fmt.Printf("Extracted Variables: %v\n", result.Variables)
} else {
    fmt.Printf("Test failed: %s\n", result.Error)
}
```

## Advanced Features

### Custom HTTP Client Configuration

```go
runner := goresttest.NewTestRunner("https://api.example.com")
// Access the underlying HTTP client if needed
// runner.executor.client.client.Timeout = 30 * time.Second
```

### Variable Interpolation

Variables can be used in:
- URLs: `/users/${user_id}`
- Headers: `Authorization: Bearer ${token}`
- Request bodies: `{"userId": "${user_id}"}`
- Assertion values: `expected: "${expected_name}"`
- File paths: `body_file: "${data_dir}/request.json"`

### Parallel Execution

```yaml
parallel: true
max_workers: 10
```

Or via CLI:
```bash
goresttest -config tests.yaml -parallel -workers 10
```

## Examples

See the [examples](examples/) directory for complete working examples:

- [library_usage.go](examples/library_usage.go) - Comprehensive library usage examples
- [api_tests.yml](examples/api_tests.yml) - Sample test configuration

## Development

### Running Tests

```bash
go test -v
```

### Building CLI

```bash
go build ./cmd/goresttest
```

## License

This project is licensed under the MIT License.