package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPClient_ExecuteRequest_BodyFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"received": "` + string(body) + `"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	bodyFile := filepath.Join(tempDir, "test_body.json")
	testContent := `{"name": "Test User", "id": ${user_id}}`

	err := os.WriteFile(bodyFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		test      Test
		variables map[string]string
		wantError bool
		checkBody bool
		expected  string
	}{
		{
			name: "body_file with variable interpolation",
			test: Test{
				Name:     "Test Body File",
				Method:   "POST",
				URL:      "/test",
				BodyFile: bodyFile,
			},
			variables: map[string]string{
				"user_id": "123",
			},
			wantError: false,
			checkBody: true,
			expected:  `{"name": "Test User", "id": 123}`,
		},
		{
			name: "body_file with non-existent file",
			test: Test{
				Name:     "Test Non-existent File",
				Method:   "POST",
				URL:      "/test",
				BodyFile: "/non/existent/file.json",
			},
			variables: map[string]string{},
			wantError: true,
		},
		{
			name: "both body and body_file specified",
			test: Test{
				Name:     "Test Both Body Fields",
				Method:   "POST",
				URL:      "/test",
				Body:     "inline body",
				BodyFile: bodyFile,
			},
			variables: map[string]string{},
			wantError: true,
		},
		{
			name: "body_file with variable in path",
			test: Test{
				Name:     "Test Variable in Path",
				Method:   "POST",
				URL:      "/test",
				BodyFile: "${temp_dir}/test_body.json",
			},
			variables: map[string]string{
				"temp_dir": tempDir,
				"user_id":  "456",
			},
			wantError: false,
			checkBody: true,
			expected:  `{"name": "Test User", "id": 456}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.ExecuteRequest(tt.test, tt.variables)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if result.Success {
					t.Errorf("Expected result.Success to be false")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Success {
				t.Errorf("Expected result.Success to be true, error: %s", result.Error)
				return
			}

			if result.StatusCode != 200 {
				t.Errorf("Expected status code 200, got %d", result.StatusCode)
			}

			if tt.checkBody && result.Response != "" {
				// The server echoes back the received body in JSON format
				if !strings.Contains(result.Response, tt.expected) {
					t.Errorf("Expected response to contain %q, got %q", tt.expected, result.Response)
				}
			}
		})
	}
}

func TestHTTPClient_ExecuteRequest_InlineBody(t *testing.T) {
	// Test that inline body still works
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write([]byte(`{"received": "` + string(body) + `"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	test := Test{
		Name:   "Test Inline Body",
		Method: "POST",
		URL:    "/test",
		Body:   `{"name": "${user_name}", "active": true}`,
	}

	variables := map[string]string{
		"user_name": "John Doe",
	}

	result, err := client.ExecuteRequest(test, variables)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected result.Success to be true, error: %s", result.Error)
	}

	if result.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", result.StatusCode)
	}

	expected := `{"name": "John Doe", "active": true}`
	if !strings.Contains(result.Response, expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, result.Response)
	}
}

func TestHTTPClient_ExecuteRequest_NoBody(t *testing.T) {
	// Test that requests without body still work
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"method": "` + r.Method + `"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	test := Test{
		Name:   "Test No Body",
		Method: "GET",
		URL:    "/test",
	}

	result, err := client.ExecuteRequest(test, map[string]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected result.Success to be true, error: %s", result.Error)
	}

	if result.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", result.StatusCode)
	}

	if !strings.Contains(result.Response, `"method": "GET"`) {
		t.Errorf("Expected response to contain method GET, got %q", result.Response)
	}
}
