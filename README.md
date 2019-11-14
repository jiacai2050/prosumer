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

There is a [quickstart example](./examples/quickstart/main.go) , view the [GoDoc](https://godoc.org/github.com/jiacai2050/prosumer) for details.

## License

[MIT License](./LICENSE)
