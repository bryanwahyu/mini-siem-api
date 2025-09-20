package pkg

import "time"

func TodayUTC() time.Time {
    now := time.Now().UTC()
    return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
