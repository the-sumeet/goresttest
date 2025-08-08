package goresttest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTTPClient handles HTTP requests for tests
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates a new HTTPClient with the specified base URL
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// ExecuteRequest executes a test request and returns the result
func (c *HTTPClient) ExecuteRequest(test Test, variables map[string]string) (*TestResult, error) {
	start := time.Now()
	
	url := c.buildURL(test.URL)
	url = InterpolateVariables(url, variables)
	
	method := strings.ToUpper(test.Method)
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if test.Body != "" && test.BodyFile != "" {
		return &TestResult{
			Name:    test.Name,
			Success: false,
			Error:   "cannot specify both 'body' and 'body_file' in the same test",
		}, fmt.Errorf("cannot specify both 'body' and 'body_file' in the same test")
	}
	
	if test.Body != "" {
		bodyStr := InterpolateVariables(test.Body, variables)
		body = bytes.NewBufferString(bodyStr)
	} else if test.BodyFile != "" {
		bodyFilePath := InterpolateVariables(test.BodyFile, variables)
		bodyContent, err := os.ReadFile(bodyFilePath)
		if err != nil {
			return &TestResult{
				Name:    test.Name,
				Success: false,
				Error:   fmt.Sprintf("failed to read body file '%s': %v", bodyFilePath, err),
			}, fmt.Errorf("failed to read body file '%s': %w", bodyFilePath, err)
		}
		bodyStr := InterpolateVariables(string(bodyContent), variables)
		body = bytes.NewBufferString(bodyStr)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &TestResult{
			Name:    test.Name,
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, err
	}

	for key, value := range test.Headers {
		interpolatedValue := InterpolateVariables(value, variables)
		req.Header.Set(key, interpolatedValue)
	}

	if test.Timeout > 0 {
		c.client.Timeout = test.Timeout
	}

	resp, err := c.client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		return &TestResult{
			Name:     test.Name,
			Success:  false,
			Duration: duration,
			Error:    fmt.Sprintf("request failed: %v", err),
		}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestResult{
			Name:       test.Name,
			Success:    false,
			StatusCode: resp.StatusCode,
			Duration:   duration,
			Headers:    resp.Header,
			Error:      fmt.Sprintf("failed to read response body: %v", err),
		}, err
	}

	result := &TestResult{
		Name:       test.Name,
		Success:    true,
		StatusCode: resp.StatusCode,
		Duration:   duration,
		Response:   string(responseBody),
		Headers:    resp.Header,
		Variables:  make(map[string]string),
	}

	return result, nil
}

func (c *HTTPClient) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	
	baseURL := strings.TrimSuffix(c.baseURL, "/")
	path = strings.TrimPrefix(path, "/")
	
	if baseURL == "" {
		return path
	}
	
	return fmt.Sprintf("%s/%s", baseURL, path)
}