package prosumer

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {

	var received []int
	consumer := func(ls []interface{}) error {
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 200

	config := DefaultConfig(Consumer(consumer))
	config.BatchSize = 21
	config.BufferSize = bufferSize
	config.NumConsumer = 1
	config.RejectPolicy = Block
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(i)
	}

	coord.Close(true)

	assert.Equal(t, maxLoop, len(received))

	for i := 0; i < maxLoop; i++ {
		assert.Equal(t, i, received[i])
	}

}

func TestDiscard(t *testing.T) {

	var received []int
	consumer := func(ls []interface{}) error {
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 200

	config := DefaultConfig(Consumer(consumer))
	config.BatchSize = 21
	config.BufferSize = bufferSize
	config.NumConsumer = 1
	config.RejectPolicy = Discard
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(i)
	}

	coord.Close(true)

	assert.True(t, maxLoop >= len(received))
	assert.True(t, bufferSize <= len(received))

	asc := true
	for i := 1; i < len(received); i++ {
		if received[i-1] > received[i] {
			asc = false
			break
		}

	}
	assert.True(t, asc)
}

func TestDiscardOldest(t *testing.T) {

	var received []int
	consumer := func(ls []interface{}) error {
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 200

	config := DefaultConfig(Consumer(consumer))
	config.BatchSize = 21
	config.BufferSize = bufferSize
	config.NumConsumer = 1
	config.RejectPolicy = DiscardOldest
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(i)
	}

	coord.Close(true)

	assert.True(t, maxLoop >= len(received))
	assert.True(t, bufferSize <= len(received))

	asc := true
	for i := 1; i < len(received); i++ {
		if received[i-1] > received[i] {
			asc = false
			break
		}

	}
	assert.True(t, asc)
}

func TestMultipleConsumer(t *testing.T) {

	var received []int
	var mux sync.Mutex
	consumer := func(ls []interface{}) error {
		mux.Lock()
		defer mux.Unlock()
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 20000

	config := DefaultConfig(Consumer(consumer))
	config.BatchSize = 21
	config.BufferSize = bufferSize
	config.NumConsumer = 20
	config.RejectPolicy = Block
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(i)
	}

	coord.Close(true)

	assert.Equal(t, len(received), maxLoop)
}

func TestBatchInterval(t *testing.T) {

	var received []int
	var mux sync.Mutex
	var err = errors.New("test")
	consumer := func(ls []interface{}) error {
		mux.Lock()
		defer mux.Unlock()
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return err
	}

	maxLoop := 10

	config := DefaultConfig(Consumer(consumer))
	config.BatchSize = 3
	config.BatchInterval = 80 * time.Millisecond
	config.BufferSize = 10
	config.NumConsumer = 1
	config.RejectPolicy = Block
	config.ErrCallback = func(ls []interface{}, e error) {
		assert.Equal(t, err, e)
	}
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(i)
		if i == 1 {
			// test consume when
			// 1. dequeue return false
			// 2. batchInterval is met
			// 3. batchSize isn't met
			time.Sleep(300 * time.Millisecond)
		} else {
			// test consume when
			// 1. dequeue return true
			// 2. batchInterval is met
			// 3. batchSize isn't met
			time.Sleep(50 * time.Millisecond)
		}
	}

	coord.Close(true)

	assert.Equal(t, len(received), maxLoop)
}
