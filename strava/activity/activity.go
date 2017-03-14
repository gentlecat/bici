package activity

import (
	"github.com/strava/go.strava"
)

func GetActivity(client *strava.Client, id int64) (*strava.ActivityDetailed, error) {
	service := strava.NewActivitiesService(client)
	return service.Get(id).
		IncludeAllEfforts().
		Do()
}
