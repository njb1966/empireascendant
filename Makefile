BIN := bin/interdoor-dominion
DB ?= var/empireascendant.db

.PHONY: help test build run smoke clean

help:
	@echo "make test   - run unit tests"
	@echo "make build  - build bin/interdoor-dominion"
	@echo "make run    - run local terminal review mode"
	@echo "make smoke  - build and run tests"
	@echo "make clean  - remove local build/runtime output"

test:
	go test ./...

build:
	mkdir -p bin
	go build -o $(BIN) ./cmd/interdoor-dominion

run: build
	mkdir -p var
	$(BIN) -db $(DB) -stdio

smoke: test build

clean:
	rm -rf bin var
