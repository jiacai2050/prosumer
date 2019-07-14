# Prosumer

[![GoDoc](https://godoc.org/github.com/jiacai2050/prosumer?status.svg)](https://godoc.org/github.com/jiacai2050/prosumer)
[![Build Status](https://travis-ci.org/jiacai2050/prosumer.svg?branch=master)](https://travis-ci.org/jiacai2050/prosumer)

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
var received []int
consumer := func(ls []interface{}) error {
	for _, e := range ls {
		received = append(received, e.(int))
	}
	return nil
}
coord := NewCoordinator(DefaultConfig(Consumer(consumer)))
coord.Start()
maxLoop := 200
for i := 0; i < maxLoop; i++ {
	coord.Put(i)
}
// Close is graceful, blocking. All elements inside buffer queue will make sure to be consumed.
coord.Close()

assert.Equal(t, maxLoop, len(received))
for i := 0; i < maxLoop; i++ {
	assert.Equal(t, i, received[i])
}

```
By default, log will print to stdout, you can custom `PROSUMER_LOGFILE` environment to redirect to a file.

## License

[MIT License](http://liujiacai.net/license/MIT.html?year=2019) © Jiacai Liu
