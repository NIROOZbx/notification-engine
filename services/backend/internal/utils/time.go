package utils

import (
	"errors"
	"time"
)

var (
	ErrInvalidDuration = errors.New("duration must be greater than zero")
	ErrDurationTooLong = errors.New("duration exceeds maximum allowed (365 days)")
)

func CalculateExpiry(days int) (*time.Time, error) {

	if days <= 0 {
		return nil, ErrInvalidDuration
	}
	if days > 365 {
		return nil, ErrDurationTooLong
	}

	expiry := time.Now().AddDate(0, 0, days)

	return &expiry, nil

}
