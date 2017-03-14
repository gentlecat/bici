package segment

import (
	"github.com/strava/go.strava"
)

// HasCompleted function determines if a segment has been attempted
// during an activity
func HasAttempted(activity *strava.ActivityDetailed, segmentID int64) bool {
	for _, effort := range activity.SegmentEfforts {
		if effort.Segment.Id == segmentID {
			return true
		}
	}
	return false
}

func HasAttemptedAny(activity *strava.ActivityDetailed, segmentIDs []int64) bool {
	for _, segmentID := range segmentIDs {
		if HasAttempted(activity, segmentID) == true {
			return true
		}
	}
	return false
}
