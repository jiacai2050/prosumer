package prosumer

import (
	"sync"
	"time"
)

// Consumer process elements from queue
type Consumer func(ls []interface{}) error

// Coordinator implements a producer-consumer workflow.
// Put() add new elements into inner buffer queue, and will be processed by Consumer.
type Coordinator struct {
	// inner buffer
	*Queue

	consumer      Consumer
	numConsumer   int
	batchSize     int
	batchInterval time.Duration
	errCallback   func(ls []interface{}, err error)

	close     chan struct{} // notify consumer to close
	waitClose *sync.WaitGroup
}

type Config struct {
	// BufferSize set inner buffer queue's size
	BufferSize int
	// RejectPolicy decide what to do when new element is added and the queue is full
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
			log.Errorf("consumer failed. list: %v, err: %v", ls, err)
		},

		NumConsumer: 2,
		Consumer:    con,
	}
}

func NewCoordinator(config Config) Coordinator {
	var waitClose sync.WaitGroup
	waitClose.Add(config.NumConsumer + 1) // +1 for queue

	c := Coordinator{
		Queue:       NewQueue(config.BufferSize, config.RejectPolicy, &waitClose),
		consumer:    config.Consumer,
		numConsumer: config.NumConsumer,

		batchSize:     config.BatchSize,
		batchInterval: config.BatchInterval,
		errCallback:   config.ErrCallback,

		close:     make(chan struct{}),
		waitClose: &waitClose,
	}
	return c
}

// Start the coordinator to consume elements from queue.
func (c Coordinator) Start() {
	for i := 0; i < c.numConsumer; i++ {
		go func(i int) {
			log.Infof("start consumer %d...", i)
			var received []interface{}
			last := time.Now()
			doWork := func() {
				if len(received) > 0 {
					err := c.consumer(received)
					last = time.Now()
					received = nil
					if err != nil {
						c.errCallback(received, err)
					}
				}
			}

			for {
				if e, ok := c.Queue.Dequeue(); ok {
					received = append(received, e)
					if len(received) >= c.batchSize {
						doWork()
					}
				} else {
					select {
					case <-c.close:
						log.Infof("consumer %d start close...", i)
						doWork()
						c.waitClose.Done()
						goto close
					case <-time.After(time.Until(last.Add(c.batchInterval))):
						doWork()
					}

				}

			}
		close:
			log.Infof("consumer %d closed.", i)
		}(i)
	}
}

// Close is graceful, blocking. All elements inside buffer queue will make sure to be consumed.
func (c Coordinator) Close() error {
	c.Queue.Close() // stop future put
	close(c.close)  // read from closed chan always return non-blocking

	c.waitClose.Wait()
	return nil
}

// Put new element into inner buffer queue.
// If queue is full, RejectPolicy will decide how to response.
func (c Coordinator) Put(e interface{}) {
	c.Queue.Enqueue(e)
}
