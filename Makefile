BINARY := tlctl

.PHONY: build run clean costbench record

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
