package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

type Test struct {
	Name     string   `json:"name"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Timeout  int      `json:"timeout,omitempty"`
}

type Request struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

type Response struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
}

type LogEntry struct {
	TestName     string                 `json:"test_name"`
	Status       string                 `json:"status"`
	Detail       string                 `json:"detail"`
	Duration     string                 `json:"duration"`
	RequestURL   string                 `json:"request_url,omitempty"`
	ResponseCode int                    `json:"response_code,omitempty"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Expected     any                    `json:"expected,omitempty"`
	Received     any                    `json:"received,omitempty"`
}

func loadTests(filename string) ([]Test, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var tests []Test
	err = json.Unmarshal(data, &tests)
	if err != nil {
		return nil, err
	}

	return tests, nil
}

const DEFAULT_TEST_TIMEOUT = 10 * time.Second

func runTest(test Test, host string, port int) LogEntry {
	start := time.Now()

	timeout := time.Duration(test.Timeout) * time.Second
	if timeout == 0 {
		timeout = DEFAULT_TEST_TIMEOUT
	}

	client := &http.Client{
		Timeout: timeout,
	}

	fullURL := fmt.Sprintf("http://%s:%d%s", host, port, test.Request.Path)

	var requestBody *bytes.Buffer
	if test.Request.Body != nil {
		serialisedBody, err := json.Marshal(test.Request.Body)
		if err != nil {
			return LogEntry{
				TestName:   test.Name,
				Status:     "FAILED",
				Detail:     fmt.Sprintf("Failed to create request body: %v", err),
				Duration:   time.Since(start).String(),
				RequestURL: fullURL,
			}
		}
		requestBody = bytes.NewBuffer(serialisedBody)
	}

	req, err := http.NewRequest(test.Request.Method, fullURL, requestBody)
	if err != nil {
		return LogEntry{
			TestName:   test.Name,
			Status:     "FAILED",
			Detail:     fmt.Sprintf("Failed to create request: %v", err),
			Duration:   time.Since(start).String(),
			RequestURL: fullURL,
		}
	}

	for key, value := range test.Request.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return LogEntry{
			TestName:   test.Name,
			Status:     "FAILED",
			Detail:     fmt.Sprintf("Failed to send request: %v", err),
			Duration:   time.Since(start).String(),
			RequestURL: fullURL,
		}
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return LogEntry{
			TestName:     test.Name,
			Status:       "FAILED",
			Detail:       fmt.Sprintf("Failed to read response body bytes: %v", err),
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
		}
	}

	var respBody map[string]interface{}
	err = json.Unmarshal(respBodyBytes, &respBody)
	if err != nil {
		return LogEntry{
			TestName:     test.Name,
			Status:       "FAILED",
			Detail:       fmt.Sprintf("Failed to deserialise resp body: %v - %s", err, string(respBodyBytes)),
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
		}
	}

	if resp.StatusCode != test.Response.StatusCode {
		return LogEntry{
			TestName:     test.Name,
			Status:       "FAILED",
			Detail:       fmt.Sprintf("Expected status %d, got %d:", test.Response.StatusCode, resp.StatusCode),
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
			ResponseBody: respBody,
		}
	}

	for key, expectedValue := range test.Response.Headers {
		actualValue := resp.Header.Get(key)
		if actualValue != expectedValue {
			return LogEntry{
				TestName:     test.Name,
				Status:       "FAILED",
				Detail:       fmt.Sprintf("Expected header %s: %s, got %s", key, expectedValue, actualValue),
				Duration:     time.Since(start).String(),
				RequestURL:   fullURL,
				ResponseCode: resp.StatusCode,
				ResponseBody: respBody,
			}
		}
	}

	if !reflect.DeepEqual(test.Response.Body, respBody) {
		return LogEntry{
			TestName:     test.Name,
			Status:       "FAILED",
			Detail:       "Body mismatch",
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
			Expected:     test.Response.Body,
			Received:     respBody,
		}
	}

	return LogEntry{
		TestName:     test.Name,
		Status:       "PASSED",
		Detail:       "Test passed successfully",
		Duration:     time.Since(start).String(),
		RequestURL:   fullURL,
		ResponseCode: resp.StatusCode,
		ResponseBody: respBody,
	}
}

func worker(tests <-chan Test, results chan<- LogEntry, wg *sync.WaitGroup, host string, port int) {
	defer wg.Done()
	for test := range tests {
		results <- runTest(test, host, port)
	}
}

func loadTestsFromDir(dirname string) ([]Test, error) {
	var allTests []Test

	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".json") {
			tests, err := loadTests(path)
			if err != nil {
				return err
			}
			allTests = append(allTests, tests...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allTests, nil
}

func main() {
	workersFlag := flag.Int("workers", 5, "Number of concurrent workers")
	hostFlag := flag.String("host", "localhost", "Host to use for the tests")
	portFlag := flag.Int("port", 3000, "Port to use for the tests")
	outputFlag := flag.String("output", "test_report.json", "Output file for the test report")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Please provide the path to a test file or directory containing test files")
		os.Exit(1)
	}

	path := flag.Args()[0]

	var tests []Test
	var err error

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		tests, err = loadTestsFromDir(path)
	} else {
		tests, err = loadTests(path)
	}

	if err != nil {
		fmt.Printf("Error loading tests: %v\n", err)
		os.Exit(1)
	}

	if len(tests) == 0 {
		fmt.Println("No tests found")
		os.Exit(1)
	}

	numWorkers := *workersFlag
	host := *hostFlag
	port := *portFlag
	testsChan := make(chan Test, len(tests))
	resultsChan := make(chan LogEntry, len(tests))

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(testsChan, resultsChan, &wg, host, port)
	}

	for _, test := range tests {
		testsChan <- test
	}
	close(testsChan)

	wg.Wait()
	close(resultsChan)

	var failedTests = 0
	var logEntries []LogEntry

	for result := range resultsChan {
		logEntries = append(logEntries, result)
		if result.Status != "PASSED" {
			failedTests += 1
			fmt.Printf("❌ %s \n", result.TestName)
		} else {
			fmt.Printf("✅ %s \n", result.TestName)
		}
	}

	fmt.Printf("Finished %d tests: %d passed, %d failed\n", len(tests), len(tests)-failedTests, failedTests)

	reportData, err := json.MarshalIndent(logEntries, "", "  ")
	if err != nil {
		fmt.Printf("Failed to generate report: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*outputFlag, reportData, 0644)
	if err != nil {
		fmt.Printf("Failed to write report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test report written to %s\n", *outputFlag)
}
