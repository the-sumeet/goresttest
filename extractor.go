package goresttest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// VariableExtractor handles extraction of variables from test responses
type VariableExtractor struct{}

// NewVariableExtractor creates a new VariableExtractor
func NewVariableExtractor() *VariableExtractor {
	return &VariableExtractor{}
}

// ExtractVariables extracts variables from a test result using the provided extraction rules
func (ve *VariableExtractor) ExtractVariables(result *TestResult, extractions map[string]string) error {
	if result.Variables == nil {
		result.Variables = make(map[string]string)
	}
	
	for varName, expression := range extractions {
		value, err := ve.extractValue(result, expression)
		if err != nil {
			return fmt.Errorf("failed to extract variable %s: %w", varName, err)
		}
		result.Variables[varName] = value
	}
	
	return nil
}

func (ve *VariableExtractor) extractValue(result *TestResult, expression string) (string, error) {
	parts := strings.SplitN(expression, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid extraction expression format: %s", expression)
	}
	
	extractorType := parts[0]
	path := parts[1]
	
	switch extractorType {
	case "json":
		return ve.extractFromJSON(result.Response, path)
	case "header":
		return ve.extractFromHeader(result.Headers, path)
	case "regex":
		return ve.extractFromRegex(result.Response, path)
	case "css":
		return ve.extractFromCSS(result.Response, path)
	case "status":
		return strconv.Itoa(result.StatusCode), nil
	case "response_time":
		return fmt.Sprintf("%.2f", float64(result.Duration.Nanoseconds())/1e6), nil
	default:
		return "", fmt.Errorf("unsupported extractor type: %s", extractorType)
	}
}

func (ve *VariableExtractor) extractFromJSON(response, path string) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}
	
	value, err := ve.getJSONPathValue(jsonData, path)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%v", value), nil
}

func (ve *VariableExtractor) extractFromHeader(headers map[string][]string, headerName string) (string, error) {
	values := headers[headerName]
	if len(values) == 0 {
		return "", fmt.Errorf("header %s not found", headerName)
	}
	return values[0], nil
}

func (ve *VariableExtractor) extractFromRegex(response, pattern string) (string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	matches := regex.FindStringSubmatch(response)
	if len(matches) == 0 {
		return "", fmt.Errorf("regex pattern did not match")
	}
	
	if len(matches) > 1 {
		return matches[1], nil
	}
	
	return matches[0], nil
}

func (ve *VariableExtractor) extractFromCSS(response, selector string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(response))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML response: %w", err)
	}
	
	selection := doc.Find(selector)
	if selection.Length() == 0 {
		return "", fmt.Errorf("CSS selector did not match any elements")
	}
	
	return strings.TrimSpace(selection.First().Text()), nil
}

func (ve *VariableExtractor) getJSONPathValue(data interface{}, path string) (interface{}, error) {
	if strings.HasPrefix(path, "[") {
		parts := []string{path}
		if dotIndex := strings.Index(path, "."); dotIndex > 0 {
			arrayPart := path[:dotIndex]
			remaining := path[dotIndex+1:]
			parts = append([]string{arrayPart}, strings.Split(remaining, ".")...)
		}
		
		return ve.processJSONPath(data, parts)
	}
	
	cleanPath := strings.TrimPrefix(path, "$.")
	if cleanPath == "" {
		return data, nil
	}
	
	parts := strings.Split(cleanPath, ".")
	return ve.processJSONPath(data, parts)
}

func (ve *VariableExtractor) processJSONPath(data interface{}, parts []string) (interface{}, error) {
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