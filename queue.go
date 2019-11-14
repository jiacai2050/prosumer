package prosumer

import (
	"context"
	"sync"
	"time"
)

type queue struct {
	ch chan Element

	rp        RejectPolicy
	waitClose *sync.WaitGroup
}

func (q *queue) enqueue(ctx context.Context, e Element) ([]Element, error) {
	select {
	case q.ch <- e:
		return nil, nil
	default:
		switch q.rp {
		case Block:
			select {
			case <-ctx.Done():
				return []Element{e}, ctx.Err()
			case q.ch <- e:
				return nil, nil
			}
		case Discard:
			return []Element{e}, ErrDiscard
		case DiscardOldest:
			var discarded []Element
			for {
				// when discard the oldest, other worker may preempt the slot,
				// so may discard more than one element.
				if v, ok := <-q.ch; ok {
					discarded = append(discarded, v)
				}
				select {
				case q.ch <- e:
					return discarded, ErrDiscardOldest
				case <-ctx.Done():
					return discarded, ctx.Err()
				default:
					// default is required for nonblocking read
				}
			}
		}
	}

	return nil, nil
}

func (q *queue) dequeue() (Element, bool) {
	select {
	case v, ok := <-q.ch:
		return v, ok
	default:
		return nil, false
	}

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
		ch:        make(chan Element, bufferSize),
		rp:        rp,
		waitClose: waitClose,
	}
}
