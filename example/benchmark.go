package main

import (
	"context"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/jiacai2050/prosumer"
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

func main() {

	var ops uint64

	consumer := func(ls []prosumer.Element) error {
		atomic.AddUint64(&ops, uint64(len(ls)))
		return nil
	}

	go func() {
		var last uint64
		for {
			current := atomic.LoadUint64(&ops)

			log.Printf("ops=%v\n", current-last)

			last = current
			time.Sleep(time.Second)
		}

	}()
	config := prosumer.NewConfig(consumer, prosumer.SetBatchSize(512), prosumer.SetNumConsumer(1))
	c := prosumer.NewCoordinator(config)
	c.Start()

	maxLoop := math.MaxInt64
	for i := 0; i < maxLoop; i++ {
		c.Put(context.TODO(), i)
	}
}
