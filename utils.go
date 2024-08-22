package main

func getTotalTestCount(testFiles []TestFile) int {
	total := 0
	for _, tf := range testFiles {
		total += len(tf.Tests)
	}
	return total
}
