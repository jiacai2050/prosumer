package prosumer

import (
	"errors"
	"log"
	"time"
)

var (
	ErrDiscard       = errors.New("discard current element")
	ErrDiscardOldest = errors.New("discard oldest element")
)

// Consumer process elements from queue
type Element interface{}
type Consumer func(lst []Element) error
type Callback func([]Element, error)

// rejectPolicy control which elements get discarded when the queue is full
type RejectPolicy int

const (
	// Block current goroutine, no elements discarded
	Block RejectPolicy = iota
	// Discard current element
	Discard
	// DiscardOldest remove the oldest to make room for new element
	DiscardOldest
)

// Config defines params for Coordinator
type Config struct {
	bufferSize int
	rp         RejectPolicy

	batchSize     int
	batchInterval time.Duration
	cb            Callback

	consumer    Consumer
	numConsumer int
}

// SetBufferSize defines inner buffer queue's size
func SetBufferSize(bufferSize int) Option {
	return func(config *Config) {
		config.bufferSize = bufferSize
	}
}

// SetRejectPolicy defines which elements get discarded when the queue is full
func SetRejectPolicy(rp RejectPolicy) Option {
	return func(config *Config) {
		config.rp = rp
	}
}

func SetBatchSize(batchSize int) Option {
	return func(config *Config) {
		config.batchSize = batchSize
	}
}

func SetBatchInterval(interval time.Duration) Option {
	return func(config *Config) {
		config.batchInterval = interval
	}
}

// SetCallback defines callback invoked with elements and err returned from consumer
func SetCallback(cb Callback) Option {
	return func(config *Config) {
		config.cb = cb
	}
}

func SetNumConsumer(numConsumer int) Option {
	return func(config *Config) {
		config.numConsumer = numConsumer
	}
}

// Option constructs a Config
type Option func(*Config)

// NewConfig returns a config with well-defined defaults
// Warn: default rejectPolicy is Block.
func NewConfig(c Consumer, opts ...Option) Config {
	conf := Config{consumer: c}

	for _, opt := range opts {
		opt(&conf)
	}

	if conf.bufferSize == 0 {
		conf.bufferSize = 10000
	}

	if conf.batchSize == 0 {
		conf.batchSize = 100
	}

	if conf.batchInterval == 0 {
		conf.batchInterval = 500 * time.Millisecond
	}

	if conf.cb == nil {
		conf.cb = func(lst []Element, err error) {
			if err != nil {
				log.Printf("consumer failed. list: %v, err: %v", lst, err)
			}
		}
	}

	if conf.numConsumer == 0 {
		conf.numConsumer = 2
	}

	return conf
}
