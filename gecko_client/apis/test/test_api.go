package test

import "errors"

// SuccessResponseTest defines the expected result of an API call that returns SuccessResponse
type SuccessResponseTest struct {
	Success bool
	Err     error
}

// GetSuccessResponseTests returns a list of possible SuccessResponseTests
func GetSuccessResponseTests() []SuccessResponseTest {
	return []SuccessResponseTest{
		{
			Success: true,
			Err:     nil,
		},
		{
			Success: false,
			Err:     nil,
		},
		{
			Err: errors.New("Non-nil error"),
		},
	}
}
