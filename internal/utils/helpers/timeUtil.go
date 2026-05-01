package helpers

import (
	"strconv"
	"time"
)

func ToUnixTimestamp(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func FromUnixTimestamp(s string) (time.Time, error) {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}