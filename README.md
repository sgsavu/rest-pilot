## rest-pilot

This is a Go-based test runner for running HTTP tests defined in JSON files.

### Features

* Loads and executes HTTP tests from JSON files.
* Runs tests defined in multiple files concurrently.
* Ensures sequential execution of tests within a single file.
* Generates a detailed test report in JSON format.

### Requirements

* Go 1.22.5 or later
* Network access to the target host and port

### Installation

1. Clone the repository:

```sh
git clone https://github.com/yourusername/your-repository.git
```

2. Navigate to the project directory:

```sh
cd your-repository
```

3. Build the application:

```sh
go build -o rest-pilot
```

### Usage

**Command-Line Options**

| Option  | Default | Description                                              |
|---------|---------|-----------------------------------------------------------|
| -target  | .       | Path to the directory or file containing test files.      |
| -workers | 1       | Number of concurrent workers to use for running tests.     |
| -host    | 127.0.0.1 | Host where the tests will be executed.                  |
| -port    | 3000    | Port where the tests will be executed.                  |
| -output  | test_report.json | Path to the output file for the test report.              |
| -no-output  | false | If enabled does not produce the test report.              |

**Running Tests**

**Test Files**

Test files should be JSON files ending with `.test.json` and follow this structure:

```json
[
  {
    "name": "Test 1",
    "timeout": 5,
    "request": {
      "method": "GET",
      "path": "/endpoint",
      "headers": {
        "Content-Type": "application/json"
      }
    },
    "response": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "key": "value"
      }
    }
  }
]
```

**Example Usage**

* To run tests from a directory:

```sh
./rest-pilot -target /path/to/tests -workers 10 -host example.com -port 8080 -output test_report.json
```

* To run tests from a specific file:

```sh
./rest-pilot -target /path/to/tests/file.test.json -workers 10 -host example.com -port 8080 -output test_report.json
```

### Output

The application generates a JSON report of the test results. This report includes details for each test, such as pass/fail status, response codes, and any discrepancies.

### Contributing

* Fork the repository.
* Create a feature branch (`git checkout -b feature/YourFeature`).
* Commit your changes (`git commit -am 'Add new feature'`).
* Push to the branch (`git push origin feature/YourFeature`).
* Create a new Pull Request.

### License

This project is licensed under the GPL-3.0 License. See the `LICENSE` file for details.
