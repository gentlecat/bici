package strava

import (
	"fmt"
	"github.com/strava/go.strava"
	"log"
	"math"
	"time"
)

const (
	RATE_LIMIT_SLEEP_TIME    = 1 * time.Minute
	SLEEP_THRESHOLD_FRACTION = 0.95
)

func RateLimitControl() {
	fraction := float64(strava.RateLimiting.FractionReached())
	if !math.IsNaN(fraction) {
		if strava.RateLimiting.Exceeded() || fraction > SLEEP_THRESHOLD_FRACTION {
			log.Println(fmt.Sprintf("Sleeping due to rate limit. Fraction reached: %#v", fraction))
			time.Sleep(RATE_LIMIT_SLEEP_TIME)
		}
	}
}
