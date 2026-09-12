package shared

import (
	"errors"
	"strconv"
	"time"

	"roci.dev/fracdex"
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

func GetRequiredStringQueryParameter(c *AuthenticatedContext, key string) (string, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return "", MissingQueryParameterErr
	}
	if value == "" {
		return "", InvalidQueryParameterErr
	}

	return value, nil
}

func GetOptionalRankQueryParameter(c *AuthenticatedContext, key string) (*string, error) {
	rank := GetOptionalStringQueryParameter(c, key)
	if rank == nil {
		return nil, nil
	}
	if _, err := fracdex.Float64Approx(*rank); err != nil {
		return nil, InvalidQueryParameterErr
	}

	return rank, nil
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

func GetRequiredTimestampQueryParameter(c *AuthenticatedContext, key string) (time.Time, error) {
	value, exists := c.GetQuery(key)
	if !exists {
		return time.Time{}, MissingQueryParameterErr
	}

	timestamp, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, InvalidQueryParameterErr
	}

	return timestamp, nil
}
