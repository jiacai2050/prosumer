package prosumer

import (
	"sync"
	"testing"

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

	coord.Close()

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

	coord.Close()

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

	coord.Close()

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

	coord.Close()

	assert.Equal(t, len(received), maxLoop)
}
