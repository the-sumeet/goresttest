package goresttest

import (
	"fmt"
	"sort"
	"sync"
)

type TestExecutor struct {
	client            *HTTPClient
	assertionEngine   *AssertionEngine
	variableExtractor *VariableExtractor
	globalVariables   map[string]string
	testResults       map[string]*TestResult
	mutex             sync.RWMutex
}

func NewTestExecutor(baseURL string) *TestExecutor {
	return &TestExecutor{
		client:            NewHTTPClient(baseURL),
		assertionEngine:   NewAssertionEngine(),
		variableExtractor: NewVariableExtractor(),
		globalVariables:   make(map[string]string),
		testResults:       make(map[string]*TestResult),
	}
}

func (te *TestExecutor) ExecuteTestSuite(suite *TestSuite) ([]*TestResult, error) {
	te.globalVariables = suite.Variables
	if te.globalVariables == nil {
		te.globalVariables = make(map[string]string)
	}

	if suite.Parallel {
		return te.executeParallel(suite.Tests, suite.MaxWorkers)
	}

	return te.executeSequential(suite.Tests)
}

func (te *TestExecutor) executeSequential(tests []Test) ([]*TestResult, error) {
	var results []*TestResult

	for _, test := range tests {
		if !te.canExecuteTest(test) {
			continue
		}

		result, err := te.executeTest(test)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		}

		te.mutex.Lock()
		te.testResults[test.Name] = result
		te.mutex.Unlock()

		results = append(results, result)
	}

	return results, nil
}

func (te *TestExecutor) executeParallel(tests []Test, maxWorkers int) ([]*TestResult, error) {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	testChan := make(chan Test, len(tests))
	resultChan := make(chan *TestResult, len(tests))

	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for test := range testChan {
				if te.canExecuteTest(test) {
					result, err := te.executeTest(test)
					if err != nil {
						result.Success = false
						result.Error = err.Error()
					}

					te.mutex.Lock()
					te.testResults[test.Name] = result
					te.mutex.Unlock()

					resultChan <- result
				}
			}
		}()
	}

	independentTests := te.getIndependentTests(tests)
	for _, test := range independentTests {
		testChan <- test
	}
	close(testChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []*TestResult
	for result := range resultChan {
		results = append(results, result)
	}

	dependentTests := te.getDependentTests(tests)
	for _, test := range dependentTests {
		if te.canExecuteTest(test) {
			result, err := te.executeTest(test)
			if err != nil {
				result.Success = false
				result.Error = err.Error()
			}

			te.mutex.Lock()
			te.testResults[test.Name] = result
			te.mutex.Unlock()

			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

func (te *TestExecutor) executeTest(test Test) (*TestResult, error) {
	currentVariables := make(map[string]string)
	te.mutex.RLock()
	for k, v := range te.globalVariables {
		currentVariables[k] = v
	}

	for _, depName := range test.DependsOn {
		if depResult, exists := te.testResults[depName]; exists && depResult.Variables != nil {
			for k, v := range depResult.Variables {
				currentVariables[k] = v
			}
		}
	}
	te.mutex.RUnlock()

	result, err := te.client.ExecuteRequest(test, currentVariables)
	if err != nil {
		return result, err
	}

	if len(test.Extract) > 0 {
		if err := te.variableExtractor.ExtractVariables(result, test.Extract); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("variable extraction failed: %v", err)
			return result, err
		}
	}

	if len(test.Assertions) > 0 {
		assertionErrors := te.assertionEngine.RunAssertions(result, test.Assertions, currentVariables)
		if len(assertionErrors) > 0 {
			result.Success = false
			result.Error = fmt.Sprintf("assertions failed: %v", assertionErrors)
		}
	}

	return result, nil
}

func (te *TestExecutor) canExecuteTest(test Test) bool {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	for _, depName := range test.DependsOn {
		if depResult, exists := te.testResults[depName]; !exists || !depResult.Success {
			return false
		}
	}
	return true
}

func (te *TestExecutor) getIndependentTests(tests []Test) []Test {
	var independent []Test
	for _, test := range tests {
		if len(test.DependsOn) == 0 {
			independent = append(independent, test)
		}
	}
	return independent
}

func (te *TestExecutor) getDependentTests(tests []Test) []Test {
	var dependent []Test
	for _, test := range tests {
		if len(test.DependsOn) > 0 {
			dependent = append(dependent, test)
		}
	}

	sort.Slice(dependent, func(i, j int) bool {
		return len(dependent[i].DependsOn) < len(dependent[j].DependsOn)
	})

	return dependent
}
