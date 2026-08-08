package shared

import (
	"strconv"
	"time"
)

func GetLogValueForRequiredStringQueryParameter(parsedString string, err error) string {
	if err != nil {
		return "INVALID"
	}
	return parsedString
}

func GetLogValueForRequiredBoolQueryParameter(parsedBool bool, err error) string {
	if err != nil {
		return "INVALID"
	}
	return strconv.FormatBool(parsedBool)
}

func GetLogValueForOptionalBoolQueryParameter(parsedBool *bool, err error) string {
	if err != nil {
		return "INVALID"
	}
	if parsedBool == nil {
		return ""
	}
	return strconv.FormatBool(*parsedBool)
}

func GetLogValueForOptionalDateQueryParameter(parsedDate *time.Time, err error) string {
	if err != nil {
		return "INVALID"
	}
	if parsedDate == nil {
		return ""
	}
	return parsedDate.Format(time.DateOnly)
}

func GetLogValueForOptionalTimestampQueryParameter(parsedTimestamp *time.Time, err error) string {
	if err != nil {
		return "INVALID"
	}
	if parsedTimestamp == nil {
		return ""
	}
	return parsedTimestamp.Format(time.RFC3339Nano)
}

func GetLogValueForOptionalStringQueryParameter(parsedString *string) string {
	if parsedString == nil {
		return ""
	}
	return *parsedString
}

func GetLogValueForOptionalInt64QueryParameter(parsedInteger *int64, err error) string {
	if err != nil {
		return "INVALID"
	}
	if parsedInteger == nil {
		return ""
	}
	return strconv.FormatInt(*parsedInteger, 10)
}
