package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/strava/go.strava"
)

type Summit struct {
	ID     int
	Name   string
	Points uint

	// Segments that are part of that summit
	Segments []*strava.SegmentSummary
}

func GetAllSummits() ([]Summit, error) {
	summits := make([]Summit, 0)

	rows, err := db.Query("SELECT id, name, points FROM summit ORDER BY id")
	if err != nil {
		return summits, err
	}
	defer rows.Close()
	for rows.Next() {
		var summit Summit
		if err := rows.Scan(&summit.ID, &summit.Name, &summit.Points); err != nil {
			return summits, err
		}
		summits = append(summits, summit)
	}
	if err := rows.Err(); err != nil {
		return summits, err
	}

	return summits, nil
}

func GetSummit(id int64) (summit Summit, err error) {
	// Getting basic info about the summit
	err = db.QueryRow("SELECT id, name, points FROM summit WHERE id = $1", id).
		Scan(&summit.ID, &summit.Name, &summit.Points)
	switch {
	case err == sql.ErrNoRows:
		// TODO: Make this error a constant in this package to be able to handle it later better
		return summit, errors.New(fmt.Sprintf("Can't find summit with id %d", id))
	case err != nil:
		return summit, err
	}

	// Getting info about segments that are part of this summit
	rows, err := db.Query(`
		SELECT data
		FROM summit_segments
		JOIN segment ON summit_segments.segment_id = segment.id
		WHERE summit_id = $1
	`, summit.ID)
	if err != nil {
		return summit, err
	}
	defer rows.Close()
	for rows.Next() {
		var segmentJSON string
		if err := rows.Scan(&segmentJSON); err != nil {
			return summit, err
		}
		var segment strava.SegmentSummary
		if err := json.Unmarshal([]byte(segmentJSON), &segment); err != nil {
			return summit, err
		}
		summit.Segments = append(summit.Segments, &segment)
	}
	if err := rows.Err(); err != nil {
		return summit, err
	}

	return summit, nil
}

type SummitEffortActivity struct {
	Summit   Summit
	Activity strava.ActivityDetailed
}

func GetSummitActivities(athleteID int64) ([]*SummitEffortActivity, error) {
	var activities = make([]*SummitEffortActivity, 0)

	// TODO: Improve this query: segment might have been attempted multiple times in one activity
	// Might want to get actual segment data
	rows, err := db.Query(`
		SELECT
			summit.id,
			summit.name,
			summit.points,
			activity.data
		FROM summit
		JOIN summit_segments ON summit_segments.summit_id = summit.id
		JOIN activity_efforts ON activity_efforts.segment_id = summit_segments.segment_id
		JOIN activity ON activity.id = activity_efforts.activity_id
		JOIN athlete ON athlete.id = activity.athlete_id
		WHERE athlete.id = $1
	`, athleteID)
	if err != nil {
		return activities, err
	}
	defer rows.Close()
	for rows.Next() {
		var sea SummitEffortActivity

		var activityJSON string
		if err := rows.Scan(&sea.Summit.ID, &sea.Summit.Name, &sea.Summit.Points, &activityJSON); err != nil {
			return activities, err
		}
		if err := json.Unmarshal([]byte(activityJSON), &sea.Activity); err != nil {
			return activities, err
		}
		activities = append(activities, &sea)
	}
	if err := rows.Err(); err != nil {
		return activities, err
	}

	return activities, nil
}

func AddSegmentToSummit(client *strava.Client, summitID int, segmentID int64) error {
	service := strava.NewSegmentsService(client)
	segmentDetails, err := service.Get(segmentID).Do()
	if err != nil {
		return err
	}
	err = AddSegment(segmentDetails.SegmentSummary)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"INSERT INTO summit_segments (summit_id, segment_id) VALUES ($1, $2)",
		summitID,
		segmentID,
	)
	if err != nil {
		return err
	}
	return nil
}
