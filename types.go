package main

import "time"

type Status string

type TestFile struct {
	Filename string
	Tests    []Test
}

type FileResult struct {
	Filename    string        `json:"filename"`
	TestResults []TestResult  `json:"test_results"`
	Status      Status        `json:"status"`
	Duration    time.Duration `json:"duration"`
}

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

type TestResult struct {
	TestName     string                 `json:"test_name"`
	Status       Status                 `json:"status"`
	Detail       string                 `json:"detail"`
	Duration     string                 `json:"duration"`
	RequestURL   string                 `json:"request_url,omitempty"`
	ResponseCode int                    `json:"response_code,omitempty"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Expected     any                    `json:"expected,omitempty"`
	Received     any                    `json:"received,omitempty"`
}
