package storage

import (
	"encoding/json"
	"github.com/strava/go.strava"
)

func SaveActivity(activity *strava.ActivityDetailed) error {
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO activity (id, athlete_id, data) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data
		`,
		activity.Id,
		activity.Athlete.Id,
		activityJSON,
	)
	if err != nil {
		return err
	}

	err = saveSegmentEfforts(activity)
	if err != nil {
		return err
	}

	return nil
}
