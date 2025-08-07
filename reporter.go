package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

type Reporter struct{}

func NewReporter() *Reporter {
	return &Reporter{}
}

func (r *Reporter) PrintConsoleReport(testResults []*TestResult) {
	fmt.Println("=== API Test Results ===")
	fmt.Println()
	
	if len(testResults) > 0 {
		r.printTestResults(testResults)
	}
	
	r.printSummary(testResults)
}

func (r *Reporter) printTestResults(results []*TestResult) {
	fmt.Println("Test Results:")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, result := range results {
		status := "✓ PASS"
		if !result.Success {
			status = "✗ FAIL"
		}
		
		fmt.Printf("%-50s %s (%v)\n", result.Name, status, result.Duration)
		
		if !result.Success && result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}
		
		if result.StatusCode > 0 {
			fmt.Printf("  Status: %d\n", result.StatusCode)
		}
		
		if len(result.Variables) > 0 {
			fmt.Printf("  Extracted variables: %v\n", result.Variables)
		}
		
		fmt.Println()
	}
}


func (r *Reporter) printSummary(testResults []*TestResult) {
	fmt.Println(strings.Repeat("=", 80))
	
	if len(testResults) > 0 {
		passed := 0
		failed := 0
		
		for _, result := range testResults {
			if result.Success {
				passed++
			} else {
				failed++
			}
		}
		
		fmt.Printf("Tests: %d total, %d passed, %d failed\n", len(testResults), passed, failed)
	}
}

func (r *Reporter) GenerateJSONReport(testResults []*TestResult, filename string) error {
	report := struct {
		Timestamp time.Time     `json:"timestamp"`
		Tests     []*TestResult `json:"tests"`
		Summary   struct {
			TotalTests  int `json:"total_tests"`
			PassedTests int `json:"passed_tests"`
			FailedTests int `json:"failed_tests"`
		} `json:"summary"`
	}{
		Timestamp: time.Now(),
		Tests:     testResults,
	}
	
	passed := 0
	failed := 0
	for _, result := range testResults {
		if result.Success {
			passed++
		} else {
			failed++
		}
	}
	
	report.Summary.TotalTests = len(testResults)
	report.Summary.PassedTests = passed
	report.Summary.FailedTests = failed
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}
	
	return os.WriteFile(filename, data, 0644)
}

func (r *Reporter) GenerateHTMLReport(testResults []*TestResult, filename string) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>API Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .summary { display: flex; gap: 20px; margin-bottom: 20px; }
        .summary-card { background: #fff; border: 1px solid #ddd; padding: 15px; border-radius: 5px; flex: 1; }
        .tests { margin-bottom: 30px; }
        .test { border: 1px solid #ddd; margin-bottom: 10px; border-radius: 5px; }
        .test-header { padding: 10px; background: #f9f9f9; cursor: pointer; }
        .test-body { padding: 10px; display: none; }
        .pass { border-left: 4px solid #4caf50; }
        .fail { border-left: 4px solid #f44336; }
        .status { font-weight: bold; }
        .pass .status { color: #4caf50; }
        .fail .status { color: #f44336; }
        .benchmarks table { width: 100%; border-collapse: collapse; }
        .benchmarks th, .benchmarks td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        .benchmarks th { background: #f5f5f5; }
        pre { background: #f5f5f5; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
    <script>
        function toggleTest(element) {
            const body = element.nextElementSibling;
            body.style.display = body.style.display === 'none' ? 'block' : 'none';
        }
    </script>
</head>
<body>
    <div class="header">
        <h1>API Test Report</h1>
        <p>Generated: {{.Timestamp}}</p>
    </div>

    <div class="summary">
        <div class="summary-card">
            <h3>Tests</h3>
            <p>Total: {{.Summary.TotalTests}}</p>
            <p>Passed: {{.Summary.PassedTests}}</p>
            <p>Failed: {{.Summary.FailedTests}}</p>
        </div>
    </div>

    {{if .Tests}}
    <div class="tests">
        <h2>Test Results</h2>
        {{range .Tests}}
        <div class="test {{if .Success}}pass{{else}}fail{{end}}">
            <div class="test-header" onclick="toggleTest(this)">
                <span class="status">{{if .Success}}✓ PASS{{else}}✗ FAIL{{end}}</span>
                <strong>{{.Name}}</strong>
                <span style="float: right;">{{.Duration}} | Status: {{.StatusCode}}</span>
            </div>
            <div class="test-body">
                {{if not .Success}}
                <p><strong>Error:</strong> {{.Error}}</p>
                {{end}}
                {{if .Variables}}
                <p><strong>Extracted Variables:</strong></p>
                <pre>{{range $key, $value := .Variables}}{{$key}}: {{$value}}
{{end}}</pre>
                {{end}}
                <p><strong>Response:</strong></p>
                <pre>{{.Response}}</pre>
            </div>
        </div>
        {{end}}
    </div>
    {{end}}

</body>
</html>`

	report := struct {
		Timestamp string        `json:"timestamp"`
		Tests     []*TestResult `json:"tests"`
		Summary   struct {
			TotalTests  int `json:"total_tests"`
			PassedTests int `json:"passed_tests"`
			FailedTests int `json:"failed_tests"`
		} `json:"summary"`
	}{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Tests:     testResults,
	}
	
	passed := 0
	failed := 0
	for _, result := range testResults {
		if result.Success {
			passed++
		} else {
			failed++
		}
	}
	
	report.Summary.TotalTests = len(testResults)
	report.Summary.PassedTests = passed
	report.Summary.FailedTests = failed
	
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()
	
	return t.Execute(file, report)
}