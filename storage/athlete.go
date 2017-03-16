package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/strava/go.strava"
)

func SaveLoginData(authResponse *strava.AuthorizationResponse) error {
	athleteJSON, err := json.Marshal(authResponse.Athlete)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO athlete (id, data, access_token) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, access_token = EXCLUDED.access_token
		`,
		authResponse.Athlete.AthleteSummary.AthleteMeta.Id,
		athleteJSON,
		authResponse.AccessToken,
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

func GetAthletesAccessToken(athleteID int64) (accessToken string, err error) {
	err = db.QueryRow("SELECT access_token FROM athlete WHERE id = $1", athleteID).Scan(&accessToken)
	switch {
	case err == sql.ErrNoRows:
		// TODO: Make this error a constant in this package to be able to handle it later better
		return accessToken, errors.New(fmt.Sprintf("Can't find athlete with id %d", athleteID))
	case err != nil:
		return accessToken, err
	}
	return accessToken, nil
}
