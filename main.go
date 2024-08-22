package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	target := flag.String("target", ".", "Target file/directory in which tests are found")
	workersFlag := flag.Int("workers", 1, "Number of concurrent workers")
	hostFlag := flag.String("host", "127.0.0.1", "Host to use for the tests")
	portFlag := flag.Int("port", 3000, "Port to use for the tests")
	outputFlag := flag.String("output", "test_report.json", "Output file for the test report")
	noOutputFlag := flag.Bool("no-output", false, "If enabled does not produce the test report.")
	flag.Parse()

	var testFiles []TestFile
	var err error

	info, err := os.Stat(*target)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		testFiles, err = scanForTestFiles(*target)
		if err != nil {
			fmt.Printf("Error scanning for test files: %v\n", err)
			os.Exit(1)
		}
	} else {
		tests, err := loadTestsFromFile(*target)
		if err != nil {
			fmt.Printf("Error loading tests: %v\n", err)
			os.Exit(1)
		}
		testFiles = []TestFile{{Filename: *target, Tests: tests}}
	}

	if len(testFiles) == 0 {
		fmt.Println("No tests found", *target)
		os.Exit(1)
	}

	totalTests := getTotalTestCount(testFiles)
	numWorkers := *workersFlag
	host := *hostFlag
	port := *portFlag
	testFilesChan := make(chan TestFile, len(testFiles))
	resultsChan := make(chan FileResult, len(testFiles))

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(testFilesChan, resultsChan, &wg, host, port)
	}

	for _, testFile := range testFiles {
		testFilesChan <- testFile
	}
	close(testFilesChan)

	wg.Wait()
	close(resultsChan)

	var failedTestFiles = 0
	var failedTests = 0

	var totalDuration time.Duration
	var fileResults []FileResult

	fmt.Println(GetRandomAsciiArt())

	for result := range resultsChan {
		fileResults = append(fileResults, result)
		totalDuration += result.Duration
		if result.Status != Passed {
			failedTestFiles += 1
			fmt.Printf("❌ %s: \n", result.Filename)
		} else {
			fmt.Printf("✅ %s: \n", result.Filename)
		}

		for _, test := range result.TestResults {
			if test.Status != Passed {
				failedTests += 1
				fmt.Printf("   • ❌ %s \n", test.TestName)
			} else {
				fmt.Printf("   • ✅ %s \n", test.TestName)
			}
		}
	}

	fmt.Printf("\nTest files: %d failed | %d passed (%d total)\n", failedTestFiles, len(testFiles)-failedTestFiles, len(testFiles))
	fmt.Printf("     Tests: %d failed | %d passed (%d total)\n", failedTests, totalTests-failedTests, totalTests)
	fmt.Printf("      Time: %s\n\n", totalDuration.String())

	if !*noOutputFlag {
		reportData, err := json.MarshalIndent(fileResults, "", "  ")
		if err != nil {
			fmt.Printf("Failed to generate report: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(*outputFlag, reportData, 0644)
		if err != nil {
			fmt.Printf("Failed to write report: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Test report written to %s\n\n", *outputFlag)
	}

	if failedTestFiles > 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
