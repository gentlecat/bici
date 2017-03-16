package strava

import (
	"errors"
	"sync"
)

// Since there's no Queue type in Go, we can just use this. Kind of hacky, but
// that's ok. See https://github.com/golang/go/wiki/SliceTricks for more info.
type activityQueue struct {
	queue []ActivityRetrievalRequest
	m     sync.Mutex
}

type ActivityRetrievalRequest struct {
	ActivityID  int64
	AccessToken string
}

func NewActivityRetrievalQueue() *activityQueue {
	return &activityQueue{
		queue: make([]ActivityRetrievalRequest, 0),
	}
}

func (q *activityQueue) Push(u ActivityRetrievalRequest) {
	q.m.Lock()
	q.queue = append(q.queue, u)
	q.m.Unlock()
}

func (q *activityQueue) Pop() (*ActivityRetrievalRequest, error) {
	q.m.Lock()
	defer q.m.Unlock()
	if len(q.queue) == 0 {
		return nil, errors.New("The queue is empty")
	}
	val := q.queue[0]
	q.queue = q.queue[1:] // Discard top element
	return &val, nil
}

func (q *activityQueue) Length() int {
	q.m.Lock()
	defer q.m.Unlock()
	return len(q.queue)
}
