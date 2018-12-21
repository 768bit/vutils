package vutils

import (
	"time"
)

type timeUtils struct{}

func (tu *timeUtils) IsDaylightSavingsTime(inTime time.Time) bool {
	_, timeOffset := inTime.Zone() //with time already in correct myLocation
	_, winterOffset := time.Date(inTime.Year(), 1, 1, 0, 0, 0, 0, inTime.Location()).Zone()
	_, summerOffset := time.Date(inTime.Year(), 7, 1, 0, 0, 0, 0, inTime.Location()).Zone()

	if winterOffset > summerOffset {
		winterOffset, summerOffset = summerOffset, winterOffset
	}

	if winterOffset != summerOffset { // the location has daylight saving
		if timeOffset != winterOffset {
			return true
		}
	}
	return false
}

func (tu *timeUtils) GetDaylightSavingsTimeOffset(inTime time.Time) time.Duration {
	_, timeOffset := inTime.Zone() //with time already in correct myLocation
	winterTime := time.Date(inTime.Year(), 1, 1, 0, 0, 0, 0, inTime.Location())
	summerTime := time.Date(inTime.Year(), 7, 1, 0, 0, 0, 0, inTime.Location())
	_, winterOffset := winterTime.Zone()
	_, summerOffset := summerTime.Zone()

	if winterOffset > summerOffset {
		winterOffset, summerOffset = summerOffset, winterOffset
	}

	if winterOffset != summerOffset { // the location has daylight saving
		if timeOffset != winterOffset {
			return summerTime.Sub(winterTime)
		}
	}
	return time.Duration(0)
}

var Time = &timeUtils{}
