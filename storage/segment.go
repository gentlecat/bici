package storage

import (
	"encoding/json"
	"github.com/strava/go.strava"
)

func saveSegmentEfforts(activity *strava.ActivityDetailed) error {
	// In case this is not the first time this activity is imported,
	// removing all known segment efforts.
	err := deleteActivityEfforts(activity.Id)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	effortInsertStmt, err := tx.Prepare("INSERT INTO activity_efforts (activity_id, segment_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	segmentInsertStmt, err := tx.Prepare(`
		INSERT INTO segment (id, data) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data
	`)
	if err != nil {
		return err
	}
	for _, effort := range activity.SegmentEfforts {

		// Inserting info about the segment in case it's missing
		segmentJSON, err := json.Marshal(effort.Segment)
		if err != nil {
			return err
		}
		if _, err = segmentInsertStmt.Exec(effort.Segment.Id, segmentJSON); err != nil {
			return err
		}

		// Inserting info about the effort itself
		if _, err = effortInsertStmt.Exec(activity.Id, effort.Segment.Id); err != nil {
			return err
		}
	}
	tx.Commit()

	return nil
}

func deleteActivityEfforts(activityID int64) error {
	_, err := db.Exec("DELETE FROM activity_efforts WHERE activity_id = $1", activityID)
	if err != nil {
		return err
	}
	return nil
}
