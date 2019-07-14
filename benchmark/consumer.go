package main

import (
	"fmt"
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

	consumer := func(ls []interface{}) error {
		atomic.AddUint64(&ops, uint64(len(ls)))
		return nil
	}

	go func() {
		var last uint64
		for {
			now := time.Now()
			current := atomic.LoadUint64(&ops)

			fmt.Printf("%s ops=%v\n", now, current-last)

			last = current
			time.Sleep(time.Second)
		}

	}()
	conf := prosumer.DefaultConfig(prosumer.Consumer(consumer))
	conf.NumConsumer = 1
	conf.BatchSize = 512
	c := prosumer.NewCoordinator(conf)
	c.Start()

	maxLoop := math.MaxInt64
	for i := 0; i < maxLoop; i++ {
		c.Put(i)
	}
}
