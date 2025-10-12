.PHONY: test bench

test:
	go test -v -race ./...

bench:
	go test -bench=. -run=^$$ ./...
