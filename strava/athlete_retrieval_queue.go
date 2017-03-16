package strava

import (
	"errors"
	"sync"
)

// Since there's no Queue type in Go, we can just use this. Kind of hacky, but
// that's ok. See https://github.com/golang/go/wiki/SliceTricks for more info.
type athleteQueue struct {
	queue []AthleteRetrievalRequest
	m     sync.Mutex
}

type AthleteRetrievalRequest struct {
	AccessToken string // Athlete's access token
}

func NewAthleteRetrievalQueue() *athleteQueue {
	return &athleteQueue{
		queue: make([]AthleteRetrievalRequest, 0),
	}
}

func (q *athleteQueue) Push(u AthleteRetrievalRequest) {
	q.m.Lock()
	q.queue = append(q.queue, u)
	q.m.Unlock()
}

func (q *athleteQueue) Pop() (*AthleteRetrievalRequest, error) {
	q.m.Lock()
	defer q.m.Unlock()
	if len(q.queue) == 0 {
		return nil, errors.New("The queue is empty")
	}
	val := q.queue[0]
	q.queue = q.queue[1:] // Discard top element
	return &val, nil
}

func (q *athleteQueue) Length() int {
	q.m.Lock()
	defer q.m.Unlock()
	return len(q.queue)
}
