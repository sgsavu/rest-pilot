package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"
)

func runTest(test Test, host string, port int) TestResult {
	start := time.Now()

	timeout := time.Duration(test.Timeout) * time.Second
	if timeout == 0 {
		timeout = DefaultTestTimeout
	}

	client := &http.Client{
		Timeout: timeout,
	}

	fullURL := fmt.Sprintf("http://%s:%d%s", host, port, test.Request.Path)

	var requestBody io.Reader
	if test.Request.Body != nil {
		serialisedBody, err := json.Marshal(test.Request.Body)
		if err != nil {
			return TestResult{
				TestName:   test.Name,
				Status:     Failed,
				Detail:     fmt.Sprintf("Failed to create request body: %v", err),
				Duration:   time.Since(start).String(),
				RequestURL: fullURL,
			}
		}
		requestBody = bytes.NewBuffer(serialisedBody)
	}

	req, err := http.NewRequest(test.Request.Method, fullURL, requestBody)
	if err != nil {
		return TestResult{
			TestName:   test.Name,
			Status:     Failed,
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
		return TestResult{
			TestName:   test.Name,
			Status:     Failed,
			Detail:     fmt.Sprintf("Failed to send request: %v", err),
			Duration:   time.Since(start).String(),
			RequestURL: fullURL,
		}
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			TestName:     test.Name,
			Status:       Failed,
			Detail:       fmt.Sprintf("Failed to read response body bytes: %v", err),
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
		}
	}

	var respBody map[string]interface{}
	err = json.Unmarshal(respBodyBytes, &respBody)
	if err != nil {
		return TestResult{
			TestName:     test.Name,
			Status:       Failed,
			Detail:       fmt.Sprintf("Failed to deserialise resp body: %v - %s", err, string(respBodyBytes)),
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
		}
	}

	if resp.StatusCode != test.Response.StatusCode {
		return TestResult{
			TestName:     test.Name,
			Status:       Failed,
			Detail:       "Status code mismatch",
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseBody: respBody,
			Expected:     test.Response.StatusCode,
			Received:     resp.StatusCode,
		}
	}

	for key, expectedValue := range test.Response.Headers {
		actualValue := resp.Header.Get(key)
		if actualValue != expectedValue {
			return TestResult{
				TestName:     test.Name,
				Status:       Failed,
				Detail:       "Header mismatch",
				Duration:     time.Since(start).String(),
				RequestURL:   fullURL,
				ResponseCode: resp.StatusCode,
				ResponseBody: respBody,
				Expected:     map[string]string{key: expectedValue},
				Received:     map[string]string{key: actualValue},
			}
		}
	}

	if !reflect.DeepEqual(test.Response.Body, respBody) {
		return TestResult{
			TestName:     test.Name,
			Status:       Failed,
			Detail:       "Body mismatch",
			Duration:     time.Since(start).String(),
			RequestURL:   fullURL,
			ResponseCode: resp.StatusCode,
			Expected:     test.Response.Body,
			Received:     respBody,
		}
	}

	return TestResult{
		TestName: test.Name,
		Status:   Passed,
		Detail:   "Test passed successfully",
		Duration: time.Since(start).String(),
	}
}

func runTestFile(testFile TestFile, host string, port int) FileResult {
	start := time.Now()
	results := make([]TestResult, 0, len(testFile.Tests))
	allPassed := true

	for _, test := range testFile.Tests {
		result := runTest(test, host, port)
		results = append(results, result)
		if result.Status != Passed {
			allPassed = false
		}
	}

	status := Passed
	if !allPassed {
		status = Failed
	}

	return FileResult{
		Filename:    testFile.Filename,
		TestResults: results,
		Status:      status,
		Duration:    time.Since(start),
	}
}

func worker(testFiles <-chan TestFile, results chan<- FileResult, wg *sync.WaitGroup, host string, port int) {
	defer wg.Done()
	for testFile := range testFiles {
		results <- runTestFile(testFile, host, port)
	}
}
