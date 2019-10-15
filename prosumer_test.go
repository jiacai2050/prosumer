package prosumer

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	go func() {
		// for debug
		// http://localhost:6060/debug/pprof/
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			panic(err)
		}
	}()
}

func TestBlock(t *testing.T) {
	var received []int
	consumer := func(ls []Element) error {
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 200

	config := NewConfig(consumer, SetBatchSize(21), SetBufferSize(bufferSize), SetNumConsumer(1), SetRejectPolicy(Block))
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(context.TODO(), i)
	}

	coord.Close(true)

	assert.Equal(t, maxLoop, len(received))

	for i := 0; i < maxLoop; i++ {
		assert.Equal(t, i, received[i])
	}

}

func TestDiscard(t *testing.T) {
	var received []int
	consumer := func(ls []Element) error {
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 200

	config := NewConfig(consumer, SetBatchSize(21), SetBufferSize(bufferSize), SetNumConsumer(1), SetRejectPolicy(Discard))
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(context.TODO(), i)
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
	consumer := func(ls []Element) error {
		for _, e := range ls {
			received = append(received, e.(int))
		}

		return nil
	}

	bufferSize := 100
	maxLoop := 200

	config := NewConfig(consumer, SetBatchSize(21), SetBufferSize(bufferSize), SetNumConsumer(1), SetRejectPolicy(DiscardOldest))
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(context.TODO(), i)
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
	var received uint32
	consumer := func(ls []Element) error {
		atomic.AddUint32(&received, uint32(len(ls)))
		return nil
	}

	bufferSize := 100
	maxLoop := 2000

	config := NewConfig(consumer, SetBatchSize(21), SetBufferSize(bufferSize), SetNumConsumer(20), SetRejectPolicy(Block))
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(context.TODO(), i)
	}

	coord.Close(true)

	assert.Equal(t, uint32(maxLoop), atomic.LoadUint32(&received))
}

func TestBatchInterval(t *testing.T) {
	var received uint32
	var err = errors.New("test")
	consumer := func(ls []Element) error {
		atomic.AddUint32(&received, uint32(len(ls)))
		return err
	}

	maxLoop := 10

	config := NewConfig(consumer, SetBatchSize(10), SetBatchInterval(100*time.Millisecond),
		SetNumConsumer(1), SetCallback(func(ls []Element, e error) {
			assert.Equal(t, err, e)
		}))
	coord := NewCoordinator(config)

	coord.Start()

	for i := 0; i < maxLoop; i++ {
		coord.Put(context.TODO(), i)

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

	assert.Equal(t, uint32(maxLoop), atomic.LoadUint32(&received))
}

func TestPutTimeout(t *testing.T) {
	consumer := func(ls []Element) error {
		return nil
	}
	config := NewConfig(consumer, SetBufferSize(1), SetBatchSize(10), SetBatchInterval(time.Hour))
	coord := NewCoordinator(config)

	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()
	ds, err := coord.Put(ctx, 1)
	assert.Nil(t, ds)
	assert.Nil(t, err)

	ds, err = coord.Put(ctx, 2)
	assert.Equal(t, 1, len(ds))
	assert.Equal(t, 2, ds[0])
	assert.Equal(t, context.DeadlineExceeded, err)
}
