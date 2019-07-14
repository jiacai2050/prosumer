package prosumer

import (
	"sync"
	"time"
)

// RejectPolicy decide what to do when Enqueue is called and the queue is full
type RejectPolicy int

const (
	// Block current goroutine
	Block RejectPolicy = iota
	// Discard current element
	Discard
	// Remove the oldest from the queue, make room for new element
	DiscardOldest
)

type Queue struct {
	ch chan interface{}

	rp        RejectPolicy
	waitClose *sync.WaitGroup
}

func (q *Queue) Enqueue(e interface{}) {
	select {
	case q.ch <- e:
	default:
		switch q.rp {
		case Block:
			q.ch <- e
		case Discard:
			log.Warningf("inner queue full(len=%d), discard: %v", q.Size(), e)
		case DiscardOldest:
			for {
				if e, ok := <-q.ch; ok {
					log.Warningf("inner queue full(len=%d), discardOldest: %v", q.Size(), e)
				}
				select {
				case q.ch <- e:
					return
				default:
					// default is required for nonblocking read
				}
			}

		}

	}
}

func (q *Queue) Dequeue() (interface{}, bool) {
	v, ok := <-q.ch
	return v, ok
}

func (q *Queue) Size() int {
	return len(q.ch)
}

func (q *Queue) Cap() int {
	return cap(q.ch)
}

func (q *Queue) Close() {
	close(q.ch)
	for {
		if q.Size() == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	q.waitClose.Done()
}

func NewQueue(bufferSize int, rp RejectPolicy, waitClose *sync.WaitGroup) *Queue {
	return &Queue{
		ch:        make(chan interface{}, bufferSize),
		rp:        rp,
		waitClose: waitClose,
	}
}
