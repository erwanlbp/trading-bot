package util

import "time"

const Day time.Duration = 24 * time.Hour

type NowFunc func() time.Time

var Now NowFunc = time.Now
