package workertest

import "fmt"

func AssertTrue(condition bool, message string) error {
	if !condition {
		return fmt.Errorf("%s", message)
	}
	return nil
}

func AssertEqual[T comparable](actual, expected T, message string) error {
	if actual != expected {
		return fmt.Errorf("%s. expected=[%v] actual=[%v]", message, expected, actual)
	}
	return nil
}
