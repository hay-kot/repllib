// http_client_example.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hay-kot/repllib"
)

// HTTPClient holds our HTTP client state
type HTTPClient struct {
	client       *http.Client
	baseURL      string
	headers      map[string]string
	lastResponse *http.Response
	lastBody     string
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}
}

func main() {
	httpClient := NewHTTPClient()

	// Create HTTP REPL with comprehensive autocomplete
	repl := repllib.New(func(input string) string {
		return httpClient.handleCommand(input)
	}).
		WithSuggestions(
			createHTTPSuggestionProvider(),
		).
		Build()

	fmt.Println("🌐 HTTP Client REPL")
	fmt.Println("===================")
	fmt.Println("Try commands like:")
	fmt.Println("  base https://api.github.com")
	fmt.Println("  get /users/octocat")
	fmt.Println("  header Authorization 'Bearer token123'")
	fmt.Println("  post /repos --data '{\"name\":\"test\"}'")
	fmt.Println("  status")
	fmt.Println()
	fmt.Println("Use TAB for autocomplete!")
	fmt.Println()

	if err := repl.Run(); err != nil {
		log.Fatal(err)
	}
}

func (h *HTTPClient) handleCommand(input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "Enter a command. Type 'help' for available commands."
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "help":
		return h.showHelp()
	case "base", "baseurl":
		return h.setBaseURL(parts)
	case "header", "headers":
		return h.handleHeaders(parts)
	case "get":
		return h.makeRequest("GET", parts)
	case "post":
		return h.makeRequest("POST", parts)
	case "put":
		return h.makeRequest("PUT", parts)
	case "patch":
		return h.makeRequest("PATCH", parts)
	case "delete":
		return h.makeRequest("DELETE", parts)
	case "head":
		return h.makeRequest("HEAD", parts)
	case "options":
		return h.makeRequest("OPTIONS", parts)
	case "status":
		return h.showStatus()
	case "body", "response":
		return h.showLastResponse()
	case "json":
		return h.formatJSON()
	case "timeout":
		return h.setTimeout(parts)
	case "clear":
		return h.clearState()
	default:
		return fmt.Sprintf("Unknown command: %s. Type 'help' for available commands.", command)
	}
}

func (h *HTTPClient) setBaseURL(parts []string) string {
	if len(parts) < 2 {
		if h.baseURL == "" {
			return "No base URL set. Usage: base <url>"
		}
		return fmt.Sprintf("Current base URL: %s", h.baseURL)
	}

	baseURL := parts[1]
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	// Validate URL
	if _, err := url.Parse(baseURL); err != nil {
		return fmt.Sprintf("Invalid URL: %s", err)
	}

	h.baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("Base URL set to: %s", h.baseURL)
}

func (h *HTTPClient) handleHeaders(parts []string) string {
	if len(parts) == 1 {
		// Show all headers
		if len(h.headers) == 0 {
			return "No headers set."
		}
		var result strings.Builder
		result.WriteString("Current headers:\n")
		for k, v := range h.headers {
			result.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		return result.String()
	}

	if len(parts) == 2 {
		// Show specific header or clear all
		if parts[1] == "clear" {
			h.headers = make(map[string]string)
			return "All headers cleared."
		}
		if val, exists := h.headers[parts[1]]; exists {
			return fmt.Sprintf("%s: %s", parts[1], val)
		}
		return fmt.Sprintf("Header '%s' not set.", parts[1])
	}

	if len(parts) >= 3 {
		// Set header
		key := parts[1]
		value := strings.Join(parts[2:], " ")
		// Remove quotes if present
		value = strings.Trim(value, "\"'")
		h.headers[key] = value
		return fmt.Sprintf("Header set: %s: %s", key, value)
	}

	return "Usage: header <name> [value] or 'header clear'"
}

func (h *HTTPClient) makeRequest(method string, parts []string) string {
	if len(parts) < 2 {
		return fmt.Sprintf("Usage: %s <path> [--data <data>] [--json]", strings.ToLower(method))
	}

	path := parts[1]
	fullURL := h.buildURL(path)

	var body io.Reader
	var contentType string

	// Parse additional arguments
	for i := 2; i < len(parts); i++ {
		switch parts[i] {
		case "--data", "-d":
			if i+1 < len(parts) {
				bodyStr := parts[i+1]
				bodyStr = strings.Trim(bodyStr, "\"'")
				body = strings.NewReader(bodyStr)
				if contentType == "" {
					contentType = "application/x-www-form-urlencoded"
				}
				i++ // Skip next part
			}
		case "--json", "-j":
			if i+1 < len(parts) {
				bodyStr := parts[i+1]
				bodyStr = strings.Trim(bodyStr, "\"'")
				body = strings.NewReader(bodyStr)
				contentType = "application/json"
				i++ // Skip next part
			} else {
				contentType = "application/json"
			}
		}
	}

	// Create request
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return fmt.Sprintf("Error creating request: %s", err)
	}

	// Add headers
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Make request
	start := time.Now()
	resp, err := h.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return fmt.Sprintf("Request failed: %s", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %s", err)
	}

	h.lastResponse = resp
	h.lastBody = string(bodyBytes)

	// Format response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s %s\n", method, fullURL))
	result.WriteString(fmt.Sprintf("Status: %s (%d)\n", resp.Status, resp.StatusCode))
	result.WriteString(fmt.Sprintf("Time: %v\n", duration))
	result.WriteString(fmt.Sprintf("Content-Length: %d bytes\n", len(bodyBytes)))

	// Show important headers
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		result.WriteString(fmt.Sprintf("Content-Type: %s\n", ct))
	}

	// Format response body
	if len(bodyBytes) > 0 {
		result.WriteString("\nResponse Body:\n")
		if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			if formatted := h.formatJSONString(string(bodyBytes)); formatted != "" {
				result.WriteString(formatted)
			} else {
				result.WriteString(string(bodyBytes))
			}
		} else {
			// Truncate long responses
			if len(bodyBytes) > 1000 {
				result.WriteString(string(bodyBytes[:1000]))
				result.WriteString(fmt.Sprintf("\n... (truncated, %d more bytes)", len(bodyBytes)-1000))
			} else {
				result.WriteString(string(bodyBytes))
			}
		}
	}

	return result.String()
}

func (h *HTTPClient) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	if h.baseURL == "" {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return "https://httpbin.org" + path // Default to httpbin for testing
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return h.baseURL + path
}

func (h *HTTPClient) showStatus() string {
	if h.lastResponse == nil {
		return "No recent requests made."
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Last Response Status: %s (%d)\n", h.lastResponse.Status, h.lastResponse.StatusCode))
	result.WriteString(fmt.Sprintf("Content-Length: %d bytes\n", len(h.lastBody)))

	result.WriteString("\nHeaders:\n")
	for k, v := range h.lastResponse.Header {
		result.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", ")))
	}

	return result.String()
}

func (h *HTTPClient) showLastResponse() string {
	if h.lastResponse == nil {
		return "No recent requests made."
	}

	if h.lastBody == "" {
		return "Response body was empty."
	}

	return h.lastBody
}

func (h *HTTPClient) formatJSON() string {
	if h.lastBody == "" {
		return "No response body to format."
	}

	formatted := h.formatJSONString(h.lastBody)
	if formatted == "" {
		return "Response body is not valid JSON."
	}

	return formatted
}

func (h *HTTPClient) formatJSONString(jsonStr string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return ""
	}

	formatted, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return ""
	}

	return string(formatted)
}

func (h *HTTPClient) setTimeout(parts []string) string {
	if len(parts) < 2 {
		return fmt.Sprintf("Current timeout: %v", h.client.Timeout)
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Sprintf("Invalid timeout value: %s", parts[1])
	}

	h.client.Timeout = time.Duration(seconds) * time.Second
	return fmt.Sprintf("Timeout set to %d seconds", seconds)
}

func (h *HTTPClient) clearState() string {
	h.headers = make(map[string]string)
	h.baseURL = ""
	h.lastResponse = nil
	h.lastBody = ""
	return "State cleared (headers, base URL, last response)."
}

func (h *HTTPClient) showHelp() string {
	return `HTTP Client Commands:
===================

Request Methods:
  get <path>              - Make GET request
  post <path> --data <d>  - Make POST request with data
  put <path> --json <j>   - Make PUT request with JSON
  patch <path>            - Make PATCH request
  delete <path>           - Make DELETE request
  head <path>             - Make HEAD request
  options <path>          - Make OPTIONS request

Configuration:
  base <url>              - Set base URL
  header <name> <value>   - Set request header
  header <name>           - Show specific header
  headers                 - Show all headers
  header clear            - Clear all headers
  timeout <seconds>       - Set request timeout

Response:
  status                  - Show last response status
  body                    - Show last response body
  json                    - Format last response as JSON
  
Utilities:
  clear                   - Clear all state
  help                    - Show this help

Examples:
  base https://api.github.com
  header Authorization "Bearer ghp_xxxx"
  get /user
  post /repos --json '{"name":"test","private":true}'`
}

func createHTTPSuggestionProvider() repllib.Suggester {
	return repllib.NewSuggestionProvider().
		// HTTP Methods
		AddFunction("get", "Make GET request").
		AddFunction("post", "Make POST request").
		AddFunction("put", "Make PUT request").
		AddFunction("patch", "Make PATCH request").
		AddFunction("delete", "Make DELETE request").
		AddFunction("head", "Make HEAD request").
		AddFunction("options", "Make OPTIONS request").

		// Configuration
		AddFunction("base", "Set base URL").
		AddFunction("baseurl", "Set base URL").
		AddFunction("header", "Manage request headers").
		AddFunction("headers", "Show all headers").
		AddFunction("timeout", "Set request timeout").

		// Response & Utilities
		AddFunction("status", "Show last response status").
		AddFunction("body", "Show last response body").
		AddFunction("response", "Show last response body").
		AddFunction("json", "Format response as JSON").
		AddFunction("clear", "Clear all state").
		AddFunction("help", "Show help information").

		// Common headers
		AddDelegate("headers", repllib.NewSuggestionProvider().
			SetMatchAny(true).
			AddIdentifier("Authorization", "Authentication header").
			AddIdentifier("Content-Type", "Content type header").
			AddIdentifier("Accept", "Accept header").
			AddIdentifier("User-Agent", "User agent header").
			AddIdentifier("X-API-Key", "API key header"),
		).

		// HTTP status codes
		AddIdentifier("200", "OK").
		AddIdentifier("201", "Created").
		AddIdentifier("400", "Bad Request").
		AddIdentifier("401", "Unauthorized").
		AddIdentifier("403", "Forbidden").
		AddIdentifier("404", "Not Found").
		AddIdentifier("500", "Internal Server Error").

		// Common options
		AddKeyword("--data").
		AddKeyword("--json").
		AddKeyword("-d").
		AddKeyword("-j").
		AddKeyword("clear").

		// Common URLs for testing
		AddIdentifier("httpbin.org", "HTTP testing service").
		AddIdentifier("jsonplaceholder.typicode.com", "Fake REST API").
		AddIdentifier("api.github.com", "GitHub API").
		AddIdentifier("localhost:3000", "Local development server").
		AddIdentifier("localhost:8080", "Local development server")
}
