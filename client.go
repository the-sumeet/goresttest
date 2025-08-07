package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *HTTPClient) ExecuteRequest(test Test, variables map[string]string) (*TestResult, error) {
	start := time.Now()
	
	url := c.buildURL(test.URL)
	url = interpolateVariables(url, variables)
	
	method := strings.ToUpper(test.Method)
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if test.Body != "" {
		bodyStr := interpolateVariables(test.Body, variables)
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
		interpolatedValue := interpolateVariables(value, variables)
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