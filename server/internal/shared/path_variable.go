package shared

import (
	"errors"
	"strconv"
)

var InvalidPathVariableErr = errors.New("invalid path variable")

func GetRequiredInt64PathVariable(c *AuthenticatedContext, key string) (int64, error) {
	integer, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || integer <= 0 {
		return 0, InvalidPathVariableErr
	}

	return integer, nil
}
