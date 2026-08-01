.PHONY: help test-go test-python contracts build-go integration verify-core

help:
	go run ./tools/projectctl help

test-go:
	go run ./tools/projectctl test-go

test-python:
	go run ./tools/projectctl test-python

contracts:
	go run ./tools/projectctl contracts

build-go:
	go run ./tools/projectctl build-go

integration:
	go run ./tools/projectctl integration

verify-core:
	go run ./tools/projectctl verify-core
