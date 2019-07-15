package prosumer

import (
	"fmt"
	"sync"
	"time"
)

// Consumer process elements from queue
type Consumer func(ls []interface{}) error

type worker struct {
	*queue
	close     chan struct{}
	waitClose *sync.WaitGroup

	consumer      Consumer
	batchSize     int
	batchInterval time.Duration
	errCallback   func(ls []interface{}, err error)
}

func (w worker) start() {
	received := make([]interface{}, 0, w.batchSize)
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
			watchdog.Reset(w.batchInterval)

			err := w.consumer(received)

			received = nil
			if err != nil {
				w.errCallback(received, err)
			}
		}
	}

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
				watchdog.Stop()
				doWork()
				w.waitClose.Done()
				return
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
	waitClose.Add(config.NumConsumer + 1) // +1 for queue

	closeCh := make(chan struct{})
	q := newQueue(config.BufferSize, config.RejectPolicy, &waitClose)
	c := Coordinator{
		queue: q,
		worker: &worker{
			queue:         q,
			close:         closeCh,
			waitClose:     &waitClose,
			consumer:      config.Consumer,
			batchSize:     config.BatchSize,
			batchInterval: config.BatchInterval,
			errCallback:   config.ErrCallback,
		},
		numConsumer: config.NumConsumer,
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
// Due to different RejectPolicy, multiple elements may be discarded before current element put successfully.
// Common usages pattern:
// discarded, err := c.Put(e)
// if err != nil {
// 	fmt.Errorf("discarded elements %+v for err %v", discarded, err)
// }
func (c Coordinator) Put(e interface{}) ([]interface{}, error) {
	return c.queue.enqueue(e)
}

// RemainingCapacity return how many elements inner buffer queue can hold.
func (c Coordinator) RemainingCapacity() int {
	return c.queue.cap() - c.queue.size()
}

type Config struct {
	// BufferSize set inner buffer queue's size
	BufferSize int
	// RejectPolicy control which elements get discarded when the queue is full
	RejectPolicy

	// BatchSize set how many elements Consumer can get from queue
	BatchSize int
	// BatchInterval set how long Consumer can get from queue
	BatchInterval time.Duration
	// ErrCallback is invoked when Consumer return error, with elements provided.
	ErrCallback func(ls []interface{}, err error)

	Consumer
	NumConsumer int
}

// DefaultConfig return a minimum config for convenience, custom specific param according to your situation.
// Warn: default RejectPolicy is Block.
func DefaultConfig(con Consumer) Config {
	return Config{
		BufferSize:   10000,
		RejectPolicy: Block,

		BatchSize:     100,
		BatchInterval: 500 * time.Millisecond,
		ErrCallback: func(ls []interface{}, err error) {
			fmt.Errorf("consumer failed. list: %v, err: %v", ls, err)
		},

		NumConsumer: 2,
		Consumer:    con,
	}
}
