BINARY := tlctl

.PHONY: build run clean costbench record generate run-kernel run-full

build:
	go build -o bin/$(BINARY) ./cmd/tlctl

run:
	go run ./cmd/tlctl --embedded

clean:
	rm -rf bin

costbench:
	@go test -v ./internal/benchmark/

record: build
	vhs demo.tape

generate:
	cd api && buf generate proto

run-kernel:
	go run ./cmd/tlctl --embedded --kernel-addr=localhost:50054

run-full:
	go run ./cmd/tlctl --kernel-addr=localhost:50054 --valkey-addr=localhost:6379 --nats-url=nats://localhost:4222
