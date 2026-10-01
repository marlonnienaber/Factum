package shared

import (
	"errors"
	"slices"
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

func GetOptionalIdQueryParameterList(c *AuthenticatedContext, key string, maxLength int) ([]int64, error) {
	values, exists := c.GetQueryArray(key)
	if !exists {
		return []int64{}, nil
	}
	if len(values) == 0 {
		return nil, InvalidQueryParameterErr
	}
	if len(values) > maxLength {
		return nil, InvalidQueryParameterErr
	}

	integers := make([]int64, 0, len(values))
	for _, value := range values {
		integer, err := strconv.ParseInt(value, 10, 64)
		if err != nil || integer <= 0 {
			return nil, InvalidQueryParameterErr
		}
		if slices.Contains(integers, integer) {
			return nil, InvalidQueryParameterErr
		}
		integers = append(integers, integer)
	}

	return integers, nil
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
