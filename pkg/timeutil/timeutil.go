package timeutil

import "time"

// NewDefaultTimestamp returns a new default timestamp matching 2006-01-02_15-04-05.000.
func NewDefaultTimestamp() string {
	return time.Now().Format("2006-01-02_15-04-05.000")
}

// YearMonthDay return a format of 2006-01-02.
func YearMonthDay() string {
	return time.Now().Format("2006-01-02")
}
