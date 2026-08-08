package shared

import "strconv"

func GetLogValueForRequiredInt64PathVariable(parsedInteger int64, err error) string {
	if err != nil {
		return "INVALID"
	}
	return strconv.FormatInt(parsedInteger, 10)
}
