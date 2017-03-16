package activity

import (
	"fmt"
	"github.com/strava/go.strava"
	"go.roman.zone/bici/storage"
	stravaWrapper "go.roman.zone/bici/strava"
	"log"
	"time"
)

const (
	WORKER_SLEEP_TIME = 4 * time.Second
)

var (
	// TODO: Check if timestamps are correct
	StartTime = time.Date(2017, time.March, 1, 1, 0, 0, 0, time.UTC)
	EndTime   = time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)

	activityRetrievalQueue = stravaWrapper.NewActivityRetrievalQueue()
	athleteRetrievalQueue  = stravaWrapper.NewAthleteRetrievalQueue()
)

func RetrieveAthlete(accessToken string) {
	athleteRetrievalQueue.Push(stravaWrapper.AthleteRetrievalRequest{
		AccessToken: accessToken,
	})
}

// ActivityDetailsRetriever is a routine that retrieves details about activities in the
// activityRetrievalQueue and save detailed data in the database.
func ActivityDetailsRetriever() {
	for {
		stravaWrapper.RateLimitControl()
		request, err := activityRetrievalQueue.Pop()
		if err != nil {
			time.Sleep(WORKER_SLEEP_TIME)
			continue
		}
		// TODO: It might be a good idea to check if activity has already been retrieved before
		activityDetails, err := getActivity(strava.NewClient(request.AccessToken), request.ActivityID)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(fmt.Sprintf("Saving activity #%d: %s", activityDetails.Id, activityDetails.Name))
		if err := storage.SaveActivity(activityDetails); err != nil {
			log.Println(err)
			continue
		}
	}
}

// AthleteRetriever is a routine that retrieves details about previous activities of
// an athlete and sends them for processing/storage.
func AthleteRetriever() {
	for {
		request, err := athleteRetrievalQueue.Pop()
		if err != nil {
			time.Sleep(WORKER_SLEEP_TIME)
			continue
		}
		log.Println(fmt.Sprintf("Getting data for athlete with token: %s", request.AccessToken))
		err = retrieveActivitiesBetween(request.AccessToken, StartTime, EndTime)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func getActivity(client *strava.Client, id int64) (*strava.ActivityDetailed, error) {
	service := strava.NewActivitiesService(client)
	return service.Get(id).
		IncludeAllEfforts().
		Do()
}

// retrieveActivitiesBetween retrieves activities for an athlete that is
// associated with provided access token.
func retrieveActivitiesBetween(accessToken string, start, end time.Time) error {
	currentPage := 1
	service := strava.NewCurrentAthleteService(strava.NewClient(accessToken))
	for {
		stravaWrapper.RateLimitControl()
		activities, err := service.ListActivities().
			Page(currentPage).
			// So apparently there's no good limit set on the number of
			// items you can retrieve in one request. Unless they can
			// handle this easily. LOL
			PerPage(150).
			// TODO: Check if type casting to `int` below is correct
			Before(int(end.Unix())).
			After(int(start.Unix())).
			Do()
		if err != nil {
			return err
		}
		if len(activities) < 1 {
			// No more activities to retrieve!
			break
		}
		// Data that we get here doesn't include info about segments.
		// So we need to get more info about it...
		for _, activity := range activities {
			activityRetrievalQueue.Push(stravaWrapper.ActivityRetrievalRequest{
				ActivityID:  activity.Id,
				AccessToken: accessToken,
			})
		}
		currentPage++
	}
	return nil
}
