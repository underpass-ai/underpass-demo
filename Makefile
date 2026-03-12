BINARY := tlctl

.PHONY: build run clean

build:
	go build -o bin/$(BINARY) ./cmd/tlctl

run:
	go run ./cmd/tlctl --valkey-addr=localhost:6379

clean:
	rm -rf bin
