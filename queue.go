package prosumer

import (
	"errors"
	"sync"
	"time"
)

// RejectPolicy control which elements get discarded when the queue is full
type RejectPolicy int

const (
	// Block current goroutine, no element discarded
	Block RejectPolicy = iota
	// Discard current element
	Discard
	// Discard the oldest to make room for new element
	DiscardOldest
)

var (
	ErrDiscard       = errors.New("discard point")
	ErrDiscardOldest = errors.New("discard oldest point")
)

type queue struct {
	ch chan interface{}

	rp        RejectPolicy
	waitClose *sync.WaitGroup
}

func (q *queue) enqueue(e interface{}) ([]interface{}, error) {
	select {
	case q.ch <- e:
		return nil, nil
	default:
		switch q.rp {
		case Block:
			q.ch <- e
			return nil, nil
		case Discard:
			return []interface{}{e}, ErrDiscard
		case DiscardOldest:
			for {
				var discarded []interface{}
				// when discard the oldest, other worker may preempt the slot,
				// so may discard more than one element.
				if v, ok := <-q.ch; ok {
					discarded = append(discarded, v)
				}
				select {
				case q.ch <- e:
					return discarded, ErrDiscardOldest
				default:
					// default is required for nonblocking read
				}
			}

		}

	}
	return nil, nil
}

func (q *queue) dequeue() (interface{}, bool) {
	v, ok := <-q.ch
	return v, ok
}

func (q *queue) size() int {
	return len(q.ch)
}

func (q *queue) cap() int {
	return cap(q.ch)
}

func (q *queue) close(graceful bool) {
	close(q.ch)
	if graceful {
		for {
			if q.size() == 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		q.waitClose.Done()
	}
}

func newQueue(bufferSize int, rp RejectPolicy, waitClose *sync.WaitGroup) *queue {
	return &queue{
		ch:        make(chan interface{}, bufferSize),
		rp:        rp,
		waitClose: waitClose,
	}
}
