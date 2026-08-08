package shared

import (
	"errors"
	"strconv"
	"time"
)

var (
	InvalidQueryParameterErr = errors.New("invalid query parameter")
	MissingQueryParameterErr = errors.New("missing query parameter")
)

func GetOptionalStringQueryParameter(c *AuthenticatedContext, key string) *string {
	value, exists := c.GetQuery(key)
	if !exists {
		return nil
	}

	return &value
}

func GetOptionalInt64QueryParameter(c *AuthenticatedContext, key string) (*int64, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return nil, nil
	}

	integer, err := strconv.ParseInt(value, 10, 64)
	if err != nil || integer <= 0 {
		return nil, InvalidQueryParameterErr
	}

	return &integer, nil
}

func GetOptionalDateQueryParameter(c *AuthenticatedContext, key string) (*time.Time, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return nil, nil
	}

	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, InvalidQueryParameterErr
	}

	return &date, nil
}

func GetOptionalTimestampQueryParameter(c *AuthenticatedContext, key string) (*time.Time, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return nil, nil
	}

	timestamp, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, InvalidQueryParameterErr
	}

	return &timestamp, nil
}

func GetOptionalBooleanQueryParameter(c *AuthenticatedContext, key string) (*bool, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return nil, nil
	}

	boolean, err := strconv.ParseBool(value)
	if err != nil {
		return nil, InvalidQueryParameterErr
	}

	return &boolean, nil
}

func GetRequiredBooleanQueryParameter(c *AuthenticatedContext, key string) (bool, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return false, MissingQueryParameterErr
	}

	boolean, err := strconv.ParseBool(value)
	if err != nil {
		return false, InvalidQueryParameterErr
	}

	return boolean, nil
}
