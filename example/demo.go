package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"
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
	maxLoop := 10
	var wg sync.WaitGroup
	wg.Add(maxLoop)
	defer wg.Wait()

	consumer := func(ls []prosumer.Element) error {
		fmt.Printf("get %+v \n", ls)
		wg.Add(-len(ls))
		return nil
	}

	config := prosumer.NewConfig(consumer, prosumer.SetBatchSize(maxLoop+1), prosumer.SetNumConsumer(1), prosumer.SetBufferSize(maxLoop),
		prosumer.SetBatchInterval(time.Second))
	c := prosumer.NewCoordinator(config)
	c.Start()

	for i := 0; i < maxLoop; i++ {
		fmt.Printf("try put %v\n", i)
		discarded, err := c.Put(context.TODO(), i)
		if err != nil {
			fmt.Errorf("discarded elements %+v for err %v", discarded, err)
			wg.Add(-len(discarded))
		}
		time.Sleep(time.Second)
	}
	c.Close(true)
}
