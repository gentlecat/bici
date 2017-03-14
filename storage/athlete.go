package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/strava/go.strava"
)

func SaveAthleteDetails(athlete strava.AthleteDetailed) error {
	athleteJSON, err := json.Marshal(athlete)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO athlete (id, data) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data
		`,
		athlete.AthleteSummary.AthleteMeta.Id,
		athleteJSON,
	)
	if err != nil {
		return err
	}
	return nil
}

func GetAthlete(id int64) (athlete strava.AthleteDetailed, err error) {
	var athleteJSON string
	err = db.QueryRow("SELECT data FROM athlete WHERE id = $1", id).Scan(&athleteJSON)
	switch {
	case err == sql.ErrNoRows:
		// TODO: Make this error a constant in this package to be able to handle it later better
		return athlete, errors.New(fmt.Sprintf("Can't find athlete with id %d", id))
	case err != nil:
		return athlete, err
	}
	err = json.Unmarshal([]byte(athleteJSON), &athlete)
	if err != nil {
		return athlete, err
	}

	return athlete, nil
}
