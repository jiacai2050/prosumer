# Prosumer

[![GoDoc](https://godoc.org/github.com/jiacai2050/prosumer?status.svg)](https://godoc.org/github.com/jiacai2050/prosumer)
[![Build Status](https://travis-ci.org/jiacai2050/prosumer.svg?branch=master)](https://travis-ci.org/jiacai2050/prosumer)
[![codecov](https://codecov.io/gh/jiacai2050/prosumer/branch/master/graph/badge.svg)](https://codecov.io/gh/jiacai2050/prosumer)

A producer-consumer solution for Golang.

## Motivation

Go is popular for its simplicity, builtin support for concurrency, light-weight goroutine. However, there are some tricks([here](https://dave.cheney.net/2013/04/30/curious-channels) and [there](https://github.com/golang/go/issues/11344)) when to coordinate among different goroutines, especially when implements [producer-consumer pattern](https://dzone.com/articles/producer-consumer-pattern) using buffered chan.

I don't want cover details here, guys who interested can check links above. But following cannot be emphasized too much:

> Close chan is a `sender -> receiver` communication, not the reverse. [#11344](https://github.com/golang/go/issues/11344#issuecomment-117862884)

## Feature

- Inner buffer queue support RejectPolicy
- Graceful close implemented

If you have any suggestions/questions, open a PR. 

## Usage

This is just a quick introduction, view the [GoDoc](https://godoc.org/github.com/jiacai2050/prosumer) for details.

Let's start with a trivial example:

```go
func main() {
	maxLoop := 10
	var wg sync.WaitGroup
	wg.Add(maxLoop)
	defer wg.Wait()
	
	consumer := func(ls []interface{}) error {
		fmt.Printf("get %+v \n", ls)
		wg.Add(-len(ls))
		return nil
	}

	conf := prosumer.DefaultConfig(prosumer.Consumer(consumer))
	c := prosumer.NewCoordinator(conf)
	c.Start()

	for i := 0; i < maxLoop; i++ {
		fmt.Printf("try put %v\n", i)
		discarded, err := c.Put(i)
		if err != nil {
			fmt.Errorf("discarded elements %+v for err %v", discarded, err)
			wg.Add(-len(discarded))
		}
		time.Sleep(time.Second)
	}
	c.Close(true)
}
```

## License

[MIT License](http://liujiacai.net/license/MIT.html?year=2019) © Jiacai Liu
