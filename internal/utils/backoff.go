package utils

import (
	"time"
)

func GetNextRetryDelay(attemptCount int) time.Duration {
	switch attemptCount {
	case 1:
		return 1 * time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 10 * time.Minute
	case 4:
		return 30 * time.Minute
	case 5:
		return 1 * time.Hour
	default:
		return 0 
	}
}
