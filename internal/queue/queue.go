package queue

import (
	"fmt"
	"sync"
)

const reconcilerQueueSize = 4096

type Queue[T any] interface {
	Add(T) error
	Close()
}

type queue[T any] struct {
	queue  chan T
	closed bool
	lock   sync.Mutex
}

func NewQueue[T any]() (Queue[T], <-chan T) {
	ch := make(chan T, reconcilerQueueSize)
	return &queue[T]{
		queue:  ch,
		closed: false,
	}, ch
}

func (q *queue[T]) Add(req T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.closed {
		return fmt.Errorf("queue channel is closed")
	}

	q.queue <- req
	return nil
}

func (q *queue[T]) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.closed = true
	close(q.queue)
}
