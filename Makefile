.PHONY: test bench

test:
	go test -v -race ./...

benchget:
	go test -bench=BenchmarkGet -run=^$$ ./...

benchset:
	go test -bench=BenchmarkSet -run=^$$ ./...