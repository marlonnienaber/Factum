package shared

import "fmt"

func GetWrapUnexpectedErrorHelper(wrapperDescription string) func(error) error {
	return func(err error) error {
		return fmt.Errorf("%s: %w", wrapperDescription, err)
	}
}
