package main

import "time"

const (
	Failed Status = "FAILED"
	Passed Status = "PASSED"
)

const DefaultTestTimeout = 10 * time.Second
