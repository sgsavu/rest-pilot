package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	timeout := time.Duration(test.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(test.Request.Method, test.Request.URL, nil)
	if err != nil {
		fmt.Printf("Test %s failed: %v\n", test.Name, err)
		return false
	}

	for key, value := range test.Request.Headers {
		req.Header.Set(key, value)
	}

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
