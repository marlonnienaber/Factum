package shared

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func NewDate(value time.Time) Date {
	return Date{
		Time: time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC),
	}
}

func ParseDate(value string) (Date, error) {
	parsedDate, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return Date{}, fmt.Errorf("parsing date failed: %w", err)
	}

	return NewDate(parsedDate), nil
}

func (date Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(date.Format(time.DateOnly))
}

func (date *Date) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("decoding date failed: %w", err)
	}

	parsedDate, err := ParseDate(value)
	if err != nil {
		return err
	}

	*date = parsedDate
	return nil
}

func (date *Date) Scan(value any) error {
	parsedTime, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan date from %T", value)
	}

	*date = NewDate(parsedTime)
	return nil
}

func (date Date) Value() (driver.Value, error) {
	return date.Format(time.DateOnly), nil
}
