package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type Test struct {
	Name     string   `json:"name"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
}

type Response struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
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

func runTest(test Test) bool {
	req, err := http.NewRequest(test.Request.Method, test.Request.URL, nil)
	if err != nil {
		fmt.Printf("Test %s failed: %v\n", test.Name, err)
		return false
	}

	for key, value := range test.Request.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Test %s failed: %v\n", test.Name, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != test.Response.StatusCode {
		fmt.Printf("Test %s failed: expected status %d, got %d\n", test.Name, test.Response.StatusCode, resp.StatusCode)
		return false
	}

	for key, expectedValue := range test.Response.Headers {
		actualValue := resp.Header.Get(key)
		if actualValue != expectedValue {
			fmt.Printf("Test %s failed: expected header %s: %s, got %s\n", test.Name, key, expectedValue, actualValue)
			return false
		}
	}

	fmt.Printf("Test %s passed\n", test.Name)
	return true
}

func worker(tests <-chan Test, results chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for test := range tests {
		results <- runTest(test)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide the path to the test file")
		os.Exit(1)
	}

	testFile := os.Args[1]
	tests, err := loadTests(testFile)
	if err != nil {
		fmt.Printf("Error loading tests: %v\n", err)
		os.Exit(1)
	}

	numWorkers := 5
	testsChan := make(chan Test, len(tests))
	resultsChan := make(chan bool, len(tests))

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(testsChan, resultsChan, &wg)
	}

	for _, test := range tests {
		testsChan <- test
	}
	close(testsChan)

	wg.Wait()
	close(resultsChan)

	allPassed := true
	for result := range resultsChan {
		if !result {
			allPassed = false
		}
	}

	if allPassed {
		fmt.Println("All tests passed")
	} else {
		fmt.Println("Some tests failed")
		os.Exit(1)
	}
}
