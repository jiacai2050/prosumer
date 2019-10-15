package prosumer

import (
	"context"
	"sync"
	"time"
)

type worker struct {
	*queue
	close     chan struct{}
	waitClose *sync.WaitGroup

	consumer      Consumer
	batchSize     int
	batchInterval time.Duration
	cb            Callback
}

func (w worker) start() {
	received := make([]Element, 0, w.batchSize)
	watchdog := time.NewTimer(w.batchInterval)

	doWork := func() {
		if len(received) > 0 {
			if !watchdog.Stop() {
				// try to drain from the channel
				select {
				case <-watchdog.C:
				default:
				}
			}

			err := w.consumer(received)
			w.cb(received, err)
			received = received[:0]
		}
		watchdog.Reset(w.batchInterval)
	}
loop:
	for {
		if e, ok := w.queue.dequeue(); ok {
			received = append(received, e)
			if len(received) >= w.batchSize {
				doWork()
			} else {
				select {
				case <-watchdog.C:
					doWork()
				default:
				}
			}
		} else {
			select {
			case <-w.close:
				doWork()
				watchdog.Stop()
				w.waitClose.Done()
				break loop
			case <-watchdog.C:
				doWork()
			}

		}
	}
}

// Coordinator implements a producer-consumer workflow.
// Put() add new elements into inner buffer queue, and will be processed by Consumer.
type Coordinator struct {
	// inner buffer
	*queue

	*worker
	numConsumer int

	close     chan struct{} // notify worker to close
	waitClose *sync.WaitGroup
}

func NewCoordinator(config Config) Coordinator {
	var waitClose sync.WaitGroup
	waitClose.Add(config.numConsumer + 1) // +1 for queue

	closeCh := make(chan struct{})
	q := newQueue(config.bufferSize, config.rp, &waitClose)
	c := Coordinator{
		queue: q,
		worker: &worker{
			queue:         q,
			close:         closeCh,
			waitClose:     &waitClose,
			consumer:      config.consumer,
			batchSize:     config.batchSize,
			batchInterval: config.batchInterval,
			cb:            config.cb,
		},
		numConsumer: config.numConsumer,
		close:       closeCh,
		waitClose:   &waitClose,
	}
	return c
}

// Start workers to consume elements from queue.
func (c Coordinator) Start() {
	for i := 0; i < c.numConsumer; i++ {
		go c.worker.start()
	}
}

// Close closes the Coordinator, no more element can be put any more.
// It can be graceful, which means:
// 1. blocking
// 2. all remaining elements in buffer queue will make sure to be consumed.
func (c Coordinator) Close(graceful bool) error {
	c.queue.close(graceful)
	close(c.close) // read from closed chan always return non-blocking
	if graceful {
		c.waitClose.Wait()
	}

	return nil
}

// Put new element into inner buffer queue. It return error when inner buffer queue is full, and elements failed putting to queue is the first return value.
// Due to different rejectPolicy, multiple elements may be discarded before current element put successfully.
// Common usages pattern:
//
//  discarded, err := c.Put(e)
//  if err != nil {
// 	  fmt.Errorf("discarded elements %+v for err %v", discarded, err)
//  }
//
func (c Coordinator) Put(ctx context.Context, e Element) ([]Element, error) {
	return c.queue.enqueue(ctx, e)
}

// RemainingCapacity return how many elements inner buffer queue can hold.
func (c Coordinator) RemainingCapacity() int {
	return c.queue.cap() - c.queue.size()
}
